#!/usr/bin/env zx

// E2E tests for miscellaneous devslot commands (version, doctor)

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
  
  const oldCwd = process.cwd()
  cd(repoPath)
  
  try {
    await $`rm -rf .git`
    await $`git init -q`
    await $`git config user.email "test@example.com"`
    await $`git config user.name "Test User"`
    await fs.writeFile('README.md', `# ${repoName}`)
    await $`git add README.md`
    await $`git commit -q -m "Initial commit"`
  } finally {
    cd(oldCwd)
  }
  
  return repoPath
}

// Version command tests
async function testVersionCommand() {
  await setupTest('version_command')
  
  const result = await $({ nothrow: true })`${devslotBinary} version`
  
  if (!result.ok) {
    fail(`Version command failed: ${result.stderr}`)
    return
  }
  
  const output = result.stdout.trim()
  
  // Should output version in format "devslot version X.Y.Z" or "devslot version dev"
  if (!output.includes('devslot version')) {
    fail(`Version output missing "devslot version", got: ${output}`)
    return
  }
  
  // Should have a version number or "dev"
  const versionMatch = output.match(/devslot version ([\d.]+|dev)/)
  if (!versionMatch) {
    fail(`Version output format incorrect: ${output}`)
    return
  }
  
  pass()
}

async function testVersionFlag() {
  await setupTest('version_flag')
  
  // Test -v flag
  const result1 = await $({ nothrow: true })`${devslotBinary} -v`
  
  if (!result1.ok) {
    fail(`Version flag -v failed: ${result1.stderr}`)
    return
  }
  
  // Test --version flag
  const result2 = await $({ nothrow: true })`${devslotBinary} --version`
  
  if (!result2.ok) {
    fail(`Version flag --version failed: ${result2.stderr}`)
    return
  }
  
  // Both should output version
  if (!result1.stdout.includes('devslot version') || !result2.stdout.includes('devslot version')) {
    fail('Version flags should output version')
    return
  }
  
  pass()
}

// Doctor command tests
async function testDoctorHealthyProject() {
  await setupTest('doctor_healthy')
  
  const repo = await createTestRepo('healthy-repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo.git
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create test-slot`
  
  const result = await $({ nothrow: true })`${devslotBinary} doctor`
  
  if (!result.ok) {
    fail(`Doctor failed on healthy project: ${result.stderr}`)
    return
  }
  
  const output = result.stdout
  
  // Should check various aspects
  if (!output.includes('Checking') || !output.includes('devslot.yaml')) {
    fail('Doctor should check devslot.yaml')
    return
  }
  
  if (!output.includes('repos')) {
    fail('Doctor should check repositories')
    return
  }
  
  if (!output.includes('slots')) {
    fail('Doctor should check slots')
    return
  }
  
  // Should report healthy status
  if (output.includes('ERROR') || output.includes('WARNING')) {
    fail('Healthy project should not have errors or warnings')
    return
  }
  
  pass()
}

async function testDoctorMissingConfig() {
  await setupTest('doctor_missing_config')
  
  // No devslot.yaml
  const result = await $({ nothrow: true })`${devslotBinary} doctor`
  
  if (result.ok) {
    fail('Doctor should fail without devslot.yaml')
    return
  }
  
  const errorOutput = result.stderr + result.stdout
  if (!errorOutput.includes('not in a devslot project')) {
    fail(`Expected "not in a devslot project" error, got: ${errorOutput}`)
    return
  }
  
  pass()
}

async function testDoctorMissingRepo() {
  await setupTest('doctor_missing_repo')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: missing.git
    url: https://example.com/missing.git
`)
  
  // Create slots directory with a slot referencing the missing repo
  await $`mkdir -p slots/test-slot`
  
  const result = await $({ nothrow: true })`${devslotBinary} doctor`
  
  // Doctor might succeed but show warnings
  const output = result.stdout + result.stderr
  
  if (!output.includes('missing.git')) {
    fail('Doctor should report missing repository')
    return
  }
  
  pass()
}

async function testDoctorCorruptedWorktree() {
  await setupTest('doctor_corrupted_worktree')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo.git
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create broken-slot`
  
  // Corrupt the worktree
  await $`rm -rf slots/broken-slot/repo.git/.git`
  
  const result = await $({ nothrow: true })`${devslotBinary} doctor`
  
  const output = result.stdout + result.stderr
  
  if (!output.includes('broken-slot') || !output.includes('repo.git')) {
    fail('Doctor should report corrupted worktree')
    return
  }
  
  pass()
}

async function testDoctorOrphanedSlot() {
  await setupTest('doctor_orphaned_slot')
  
  const repo1 = await createTestRepo('repo1')
  const repo2 = await createTestRepo('repo2')
  
  // Start with two repos
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo1.git
    url: ${repo1}
  - name: repo2.git
    url: ${repo2}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create multi-slot`
  
  // Remove repo2 from config
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo1.git
    url: ${repo1}
`)
  
  // Remove repo2 from repos directory to simulate inconsistency
  await $`rm -rf repos/repo2.git`
  
  const result = await $({ nothrow: true })`${devslotBinary} doctor`
  
  const output = result.stdout + result.stderr
  
  // Should detect the orphaned worktree
  if (!output.includes('repo2.git') || !output.includes('multi-slot')) {
    fail('Doctor should detect orphaned worktree')
    return
  }
  
  pass()
}

async function testDoctorEmptyProject() {
  await setupTest('doctor_empty_project')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories: []
`)
  
  const result = await $({ nothrow: true })`${devslotBinary} doctor`
  
  if (!result.ok) {
    fail(`Doctor failed on empty project: ${result.stderr}`)
    return
  }
  
  const output = result.stdout
  
  // Should handle empty project gracefully
  if (output.includes('ERROR')) {
    fail('Empty project should not have errors')
    return
  }
  
  pass()
}

// Main test runner
async function runTests() {
  echo('Running devslot miscellaneous commands E2E tests...')
  echo('================================\n')
  
  // Check if binary exists
  if (!await fs.pathExists(devslotBinary)) {
    echo(chalk.red('Error: devslot binary not found. Run "make build" first.'))
    process.exit(1)
  }
  
  // Run version tests
  echo(chalk.blue('\n--- Version Command Tests ---'))
  await testVersionCommand()
  await testVersionFlag()
  
  // Run doctor tests
  echo(chalk.blue('\n--- Doctor Command Tests ---'))
  await testDoctorHealthyProject()
  await testDoctorMissingConfig()
  await testDoctorMissingRepo()
  await testDoctorCorruptedWorktree()
  await testDoctorOrphanedSlot()
  await testDoctorEmptyProject()
  
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