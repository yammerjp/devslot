#!/usr/bin/env zx

// E2E tests for devslot destroy command

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

// Tests
async function testBasicDestroy() {
  await setupTest('basic_destroy')
  
  // Setup project
  const repo = await createTestRepo('repo1')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create test-slot`
  
  // Verify slot exists
  if (!await fs.pathExists('slots/test-slot')) {
    fail('Slot was not created properly')
    return
  }
  
  // Destroy the slot
  const result = await $({ nothrow: true })`${devslotBinary} destroy test-slot`
  
  if (!result.ok) {
    fail(`Destroy failed: ${result.stderr}`)
    return
  }
  
  // Verify slot is removed
  if (await fs.pathExists('slots/test-slot')) {
    fail('Slot directory still exists after destroy')
    return
  }
  
  // Verify bare repo still exists
  if (!await fs.pathExists('repos/repo.git')) {
    fail('Bare repository was removed (should be preserved)')
    return
  }
  
  if (!result.stdout.includes('destroyed successfully')) {
    fail('Output missing success message')
    return
  }
  
  pass()
}

async function testDestroyNonExistent() {
  await setupTest('destroy_nonexistent')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories: []
`)
  
  const result = await $({ nothrow: true })`${devslotBinary} destroy nonexistent-slot`
  
  if (result.ok) {
    fail('Destroying non-existent slot should fail')
    return
  }
  
  const errorOutput = result.stderr + result.stdout
  if (!errorOutput.includes('does not exist')) {
    fail(`Expected "does not exist" error, got: ${errorOutput}`)
    return
  }
  
  pass()
}

async function testDestroyWithChanges() {
  await setupTest('destroy_with_changes')
  
  const repo = await createTestRepo('repo-with-changes')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create work-slot`
  
  // Make changes in the worktree
  await fs.writeFile('slots/work-slot/repo/new-file.txt', 'Some changes')
  
  // Destroy should still work (pre-destroy hook might warn but not block)
  const result = await $({ nothrow: true })`${devslotBinary} destroy work-slot`
  
  if (!result.ok) {
    fail(`Destroy with changes failed: ${result.stderr}`)
    return
  }
  
  if (await fs.pathExists('slots/work-slot')) {
    fail('Slot not removed')
    return
  }
  
  pass()
}

async function testDestroyWithHook() {
  await setupTest('destroy_with_hook')
  
  const repo = await createTestRepo('repo-hook')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create pre-destroy hook
  await $`mkdir -p hooks`
  await fs.writeFile('hooks/pre-destroy', `#!/bin/bash
echo "Pre-destroy hook executed for slot: $DEVSLOT_SLOT_NAME"
echo "Creating backup marker..."
mkdir -p backups
echo "$DEVSLOT_SLOT_NAME" > backups/destroyed-slot.txt
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create hook-test-slot`
  
  // Destroy with hook
  const result = await $({ nothrow: true })`${devslotBinary} destroy hook-test-slot`
  
  if (!result.ok) {
    fail(`Destroy with hook failed: ${result.stderr}`)
    return
  }
  
  // Verify hook was executed
  if (!await fs.pathExists('backups/destroyed-slot.txt')) {
    fail('Pre-destroy hook was not executed')
    return
  }
  
  const backupContent = await fs.readFile('backups/destroyed-slot.txt', 'utf-8')
  if (!backupContent.includes('hook-test-slot')) {
    fail('Hook did not receive correct slot name')
    return
  }
  
  pass()
}

async function testDestroyFailingHook() {
  await setupTest('destroy_failing_hook')
  
  const repo = await createTestRepo('repo-failing-hook')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create failing pre-destroy hook
  await $`mkdir -p hooks`
  await fs.writeFile('hooks/pre-destroy', `#!/bin/bash
echo "Pre-destroy hook preventing destruction!"
exit 1
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create protected-slot`
  
  // Try to destroy - should fail
  const result = await $({ nothrow: true })`${devslotBinary} destroy protected-slot`
  
  if (result.ok) {
    fail('Destroy should fail when pre-destroy hook fails')
    return
  }
  
  // Verify slot still exists
  if (!await fs.pathExists('slots/protected-slot')) {
    fail('Slot was removed despite hook failure')
    return
  }
  
  const errorOutput = result.stderr + result.stdout
  if (!errorOutput.includes('pre-destroy hook failed')) {
    fail('Error message should mention hook failure')
    return
  }
  
  pass()
}

async function testDestroyMultipleRepos() {
  await setupTest('destroy_multiple_repos')
  
  const repo1 = await createTestRepo('multi-repo1')
  const repo2 = await createTestRepo('multi-repo2')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo1
    url: ${repo1}
  - name: repo2
    url: ${repo2}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create multi-slot`
  
  // Verify both worktrees exist
  if (!await fs.pathExists('slots/multi-slot/repo1') || 
      !await fs.pathExists('slots/multi-slot/repo2')) {
    fail('Multi-repo slot not created properly')
    return
  }
  
  // Destroy
  const result = await $({ nothrow: true })`${devslotBinary} destroy multi-slot`
  
  if (!result.ok) {
    fail(`Destroy multi-repo slot failed: ${result.stderr}`)
    return
  }
  
  // Verify entire slot directory is removed
  if (await fs.pathExists('slots/multi-slot')) {
    fail('Slot directory not completely removed')
    return
  }
  
  // Verify bare repos still exist
  if (!await fs.pathExists('repos/repo1.git') || !await fs.pathExists('repos/repo2.git')) {
    fail('Bare repositories were removed')
    return
  }
  
  pass()
}

// Main test runner
async function runTests() {
  echo('Running devslot destroy E2E tests...')
  echo('================================\n')
  
  // Check if binary exists
  if (!await fs.pathExists(devslotBinary)) {
    echo(chalk.red('Error: devslot binary not found. Run "make build" first.'))
    process.exit(1)
  }
  
  // Run all tests
  await testBasicDestroy()
  await testDestroyNonExistent()
  await testDestroyWithChanges()
  await testDestroyWithHook()
  await testDestroyFailingHook()
  await testDestroyMultipleRepos()
  
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