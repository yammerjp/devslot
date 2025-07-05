#!/usr/bin/env zx

// E2E tests for devslot init command

// Disable verbose output for cleaner test results
$.verbose = false

// Test utilities
const testDir = (await $`mktemp -d`).stdout.trim()
const projectRoot = process.cwd()
const devslotBinary = path.join(projectRoot, 'devslot')

// Cleanup on exit
process.on('exit', () => {
  try {
    fs.rmSync(testDir, { recursive: true, force: true })
  } catch {}
})

// Test result tracking
let testsRun = 0
let testsPassed = 0

// Helper functions
async function setupTest(testName) {
  echo(chalk.yellow(`\nRunning: ${testName}`))
  testsRun++
  
  const testProjectDir = path.join(testDir, testName)
  await $`mkdir -p ${testProjectDir}`
  cd(testProjectDir)
  
  return testProjectDir
}

function pass() {
  echo(chalk.green('✓ Test passed'))
  testsPassed++
}

function fail(message) {
  echo(chalk.red(`✗ ${message}`))
}

async function createTestRepo(repoName) {
  const repoPath = path.join(testDir, 'test-repos', repoName)
  await $`mkdir -p ${repoPath}`
  
  // Initialize git repo with proper working directory
  const oldCwd = process.cwd()
  cd(repoPath)
  
  try {
    // Remove .git directory if it exists (in case we're in a git repo)
    await $`rm -rf .git`
    await $`git init -q`
    await $`git config user.email "test@example.com"`
    await $`git config user.name "Test User"`
    await fs.writeFile('README.md', `# ${repoName}`)
    await $`git add README.md`
    await $`git commit -q -m "Initial commit"`
  } catch (e) {
    echo(chalk.red(`Failed to create test repo at ${repoPath}:`))
    echo(chalk.red(e.stderr || e.stdout || e.message))
    // Check current directory status
    await $`pwd`
    await $`ls -la`
    throw e
  }
  
  cd(oldCwd)
  
  return repoPath
}

// Tests
async function testBasicInit() {
  await setupTest('basic_init')
  
  // Create test repositories
  const repo1 = await createTestRepo('repo1')
  const repo2 = await createTestRepo('repo2')
  
  // Create devslot.yaml
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - ${repo1}
  - ${repo2}
`)
  
  // Run init
  const result = await $({ nothrow: true })`${devslotBinary} init`
  
  // Assertions
  if (!result.ok) {
    fail(`Init failed: ${result.stderr}`)
    return
  }
  
  if (!await fs.pathExists('repos/repo1.git')) {
    fail('repos/repo1.git does not exist')
    return
  }
  
  if (!await fs.pathExists('repos/repo2.git')) {
    fail('repos/repo2.git does not exist')
    return
  }
  
  // Verify they are bare repositories
  const bareCheck = await $({ nothrow: true })`cd repos/repo1.git && git config --get core.bare`
  if (bareCheck.stdout.trim() !== 'true') {
    fail('repo1.git is not a bare repository')
    return
  }
  
  if (!result.stdout.includes('Cloning') || !result.stdout.includes('Init completed successfully')) {
    fail('Output missing expected messages')
    return
  }
  
  pass()
}

async function testSkipExisting() {
  await setupTest('skip_existing')
  
  const repo1 = await createTestRepo('repo1')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - ${repo1}
`)
  
  // First init
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Add marker file
  await fs.writeFile('repos/repo1.git/MARKER', 'test')
  
  // Second init
  const result = await $({ nothrow: true })`${devslotBinary} init`
  
  if (!await fs.pathExists('repos/repo1.git/MARKER')) {
    fail('Marker file was removed - repository was replaced')
    return
  }
  
  if (!result.stdout.includes('already exists, skipping')) {
    fail('Output missing skip message')
    return
  }
  
  pass()
}

async function testAllowDelete() {
  await setupTest('allow_delete')
  
  const repo1 = await createTestRepo('repo1')
  const repo2 = await createTestRepo('repo2')
  
  // Initial config with both repos
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - ${repo1}
  - ${repo2}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Update config to only have repo1
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - ${repo1}
`)
  
  // Run with --allow-delete
  const result = await $({ nothrow: true })`${devslotBinary} init --allow-delete`
  
  if (!await fs.pathExists('repos/repo1.git')) {
    fail('repo1.git was removed incorrectly')
    return
  }
  
  if (await fs.pathExists('repos/repo2.git')) {
    fail('repo2.git was not removed')
    return
  }
  
  if (!result.stdout.includes('Removing unlisted repository')) {
    fail('Output missing removal message')
    return
  }
  
  pass()
}

async function testUrlFormats() {
  await setupTest('url_formats')
  
  const repo = await createTestRepo('myrepo')
  
  // Create relative repo
  await createTestRepo('relative-repo')
  await $`mv ${testDir}/test-repos/relative-repo ./relative-repo`
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - ${repo}
  - file://${repo}
  - ./relative-repo
`)
  
  const result = await $({ nothrow: true })`${devslotBinary} init`
  
  if (!result.ok) {
    fail(`Init failed: ${result.stderr}`)
    return
  }
  
  if (!await fs.pathExists('repos/myrepo.git')) {
    fail('repos/myrepo.git does not exist')
    return
  }
  
  if (!await fs.pathExists('repos/relative-repo.git')) {
    fail('repos/relative-repo.git does not exist')
    return
  }
  
  pass()
}

async function testNoConfig() {
  await setupTest('no_config')
  
  // No devslot.yaml created
  const result = await $({ nothrow: true })`${devslotBinary} init`
  
  if (result.ok) {
    fail('Init should have failed without devslot.yaml')
    echo(`Unexpected success, output: ${result.stdout}`)
    return
  }
  
  
  // Check both stderr and stdout for the error message
  const output = result.stderr + result.stdout
  if (!output.includes('devslot.yaml not found')) {
    fail(`Expected error about missing devslot.yaml`)
    return
  }
  
  pass()
}

async function testConcurrentLock() {
  await setupTest('concurrent_lock')
  
  // Create multiple repositories to make init take longer
  const repos = []
  for (let i = 0; i < 10; i++) {
    repos.push(await createTestRepo(`lock-test-repo-${i}`))
  }
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
${repos.map(r => `  - ${r}`).join('\n')}
`)
  
  // Start many init commands in parallel to increase chance of lock conflict
  const initCommands = []
  for (let i = 0; i < 5; i++) {
    initCommands.push($({ nothrow: true })`${devslotBinary} init`)
  }
  
  // Wait for all to complete
  const results = await Promise.all(initCommands)
  
  // One should succeed, others should fail with lock error
  let successCount = 0
  let lockFailureCount = 0
  
  for (const result of results) {
    if (result.ok) {
      successCount++
    } else {
      const output = result.stderr + result.stdout
      if (output.includes('another devslot process is already running')) {
        lockFailureCount++
      }
    }
  }
  
  if (successCount !== 1 || lockFailureCount !== 4) {
    fail(`Expected 1 success and 4 lock failures, got ${successCount} successes and ${lockFailureCount} lock failures`)
    return
  }
  
  pass()
}

async function testCreateReposDir() {
  await setupTest('create_repos_dir')
  
  const repo = await createTestRepo('repo1')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - ${repo}
`)
  
  // Ensure repos directory doesn't exist
  await $`rm -rf repos`
  
  const result = await $({ nothrow: true })`${devslotBinary} init`
  
  if (!result.ok) {
    fail(`Init failed: ${result.stderr}`)
    return
  }
  
  if (!await fs.pathExists('repos')) {
    fail('repos directory was not created')
    return
  }
  
  if (!await fs.pathExists('repos/repo1.git')) {
    fail('repos/repo1.git does not exist')
    return
  }
  
  pass()
}

// Main test runner
async function runTests() {
  echo('Running devslot init E2E tests...')
  echo('================================\n')
  
  // Check if binary exists
  if (!await fs.pathExists(devslotBinary)) {
    echo(chalk.red('Error: devslot binary not found. Run "make build" first.'))
    process.exit(1)
  }
  
  // Run all tests
  await testBasicInit()
  await testSkipExisting()
  await testAllowDelete()
  await testUrlFormats()
  await testNoConfig()
  await testConcurrentLock()
  await testCreateReposDir()
  
  // Summary
  echo('\n================================')
  echo(`Tests run: ${testsRun}`)
  echo(`Tests passed: ${chalk.green(testsPassed)}`)
  
  if (testsPassed === testsRun) {
    echo(chalk.green('All tests passed!'))
    process.exit(0)
  } else {
    echo(chalk.red('Some tests failed!'))
    process.exit(1)
  }
}

// Run tests
await runTests()