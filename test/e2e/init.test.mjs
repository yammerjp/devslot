#!/usr/bin/env zx

// E2E tests for devslot init command

// Disable verbose output for cleaner test results
$.verbose = false

// Test utilities
const testDir = (await $`mktemp -d`).stdout.trim()
const projectRoot = process.cwd()
const devslotBinary = path.join(projectRoot, 'build', 'devslot')

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
  - name: repo1.git
    url: ${repo1}
  - name: repo2.git
    url: ${repo2}
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
  
  if (!result.stdout.includes('Cloning') || !result.stdout.includes('Initialization complete!')) {
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
  - name: repo1.git
    url: ${repo1}
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
  - name: repo1.git
    url: ${repo1}
  - name: repo2.git
    url: ${repo2}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Update config to only have repo1
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo1.git
    url: ${repo1}
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
  - name: test-repo.git
    url: ${repo}
  - name: test-repo-file.git
    url: file://${repo}
  - name: relative-repo.git
    url: ./relative-repo
`)
  
  const result = await $({ nothrow: true })`${devslotBinary} init`
  
  if (!result.ok) {
    fail(`Init failed: ${result.stderr}`)
    return
  }
  
  if (!await fs.pathExists('repos/test-repo.git')) {
    fail('repos/test-repo.git does not exist')
    return
  }
  
  if (!await fs.pathExists('repos/test-repo-file.git')) {
    fail('repos/test-repo-file.git does not exist')
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
  
  // Check for the error message (it appears the error is printed to stdout with exit code)
  const output = result.stderr + result.stdout
  if (!output.includes('devslot.yaml not found') && !output.includes('not in a devslot project')) {
    fail(`Expected error about missing devslot.yaml, got: ${output}`)
    return
  }
  
  pass()
}

async function testConcurrentLock() {
  await setupTest('concurrent_lock')
  
  // Create a simple test repo
  const repo1 = await createTestRepo('concurrent-test-repo')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: test-repo.git
    url: ${repo1}
`)
  
  // Start init commands in parallel with test delay
  const initCommands = []
  
  // Set environment variable to add delay in init command
  const env = { ...process.env, DEVSLOT_TEST_INIT_DELAY: '500ms' }
  
  // Start all commands at once
  for (let i = 0; i < 5; i++) {
    initCommands.push($({ nothrow: true, env })`${devslotBinary} init`)
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
  
  // Verify that we have the expected lock behavior
  // With the delay, we should have exactly 1 success and 4 lock failures
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
  - name: large-repo.git
    url: ${repo}
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
  
  if (!await fs.pathExists('repos/large-repo.git')) {
    fail('repos/large-repo.git does not exist')
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