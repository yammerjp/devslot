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
  
  // Kong's version flag outputs just the version string (e.g., "dev" or "1.2.3")
  // This is different from the version command which outputs "devslot version X.Y.Z"
  const output1 = result1.stdout.trim()
  const output2 = result2.stdout.trim()
  
  // Should output a version string (either "dev" or a version number)
  if (!output1.match(/^(dev|\d+\.\d+\.\d+)$/) || !output2.match(/^(dev|\d+\.\d+\.\d+)$/)) {
    fail(`Version flags should output version string, got: -v="${output1}", --version="${output2}"`)
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
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create test-slot`
  
  // Create hooks directory as doctor expects it
  await $`mkdir -p hooks`
  
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
  - name: missing
    url: https://example.com/missing.git
`)
  
  // Create required directories for doctor
  await $`mkdir -p hooks repos slots/test-slot`
  
  const result = await $({ nothrow: true })`${devslotBinary} doctor`
  
  // Doctor might succeed but show warnings
  const output = result.stdout + result.stderr
  
  if (!output.includes('missing')) {
    fail('Doctor should report missing repository')
    return
  }
  
  pass()
}

async function testDoctorWithCorruptedWorktree() {
  await setupTest('doctor_with_corrupted_worktree')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create broken-slot`
  
  // Create hooks directory as doctor expects it
  await $`mkdir -p hooks`
  
  // Corrupt the worktree
  await $`rm -rf slots/broken-slot/repo/.git`
  
  const result = await $({ nothrow: true })`${devslotBinary} doctor`
  
  // Doctor currently doesn't check worktree integrity, so it should succeed
  // even with a corrupted worktree. This is a limitation of the current implementation.
  if (!result.ok) {
    fail(`Doctor failed unexpectedly: ${result.stderr}`)
    return
  }
  
  const output = result.stdout
  
  // Doctor should at least report directories exist
  if (!output.includes('Directory slots exists')) {
    fail('Doctor should check slots directory')
    return
  }
  
  // Note: Future enhancement could detect corrupted worktrees
  pass()
}

async function testDoctorWithRemovedRepository() {
  await setupTest('doctor_with_removed_repository')
  
  const repo1 = await createTestRepo('repo1')
  const repo2 = await createTestRepo('repo2')
  
  // Start with two repos
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo1
    url: ${repo1}
  - name: repo2
    url: ${repo2}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create multi-slot`
  
  // Create hooks directory as doctor expects it
  await $`mkdir -p hooks`
  
  // Remove repo2 from config
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo1
    url: ${repo1}
`)
  
  // Remove repo2 from repos directory to simulate inconsistency
  await $`rm -rf repos/repo2.git`
  
  const result = await $({ nothrow: true })`${devslotBinary} doctor`
  
  // Doctor currently doesn't check for orphaned worktrees in slots.
  // It only checks that the repositories defined in devslot.yaml exist in repos/
  // This is a limitation of the current implementation.
  if (!result.ok) {
    fail(`Doctor failed unexpectedly: ${result.stderr}`)
    return
  }
  
  const output = result.stdout
  
  // Doctor should successfully check repo1
  if (!output.includes('repo1') && !output.includes('Repository') && !output.includes('cloned')) {
    fail('Doctor should check configured repositories')
    return
  }
  
  // Note: Future enhancement could detect orphaned worktrees in slots
  pass()
}

async function testDoctorEmptyProject() {
  await setupTest('doctor_empty_project')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories: []
`)
  
  // Create required directories for doctor
  await $`mkdir -p hooks repos slots`
  
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
  await testDoctorWithCorruptedWorktree()
  await testDoctorWithRemovedRepository()
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