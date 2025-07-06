#!/usr/bin/env zx

// E2E tests for devslot reload command

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
async function testBasicReload() {
  await setupTest('basic_reload')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create test-slot`
  
  // Reload the slot
  const result = await $({ nothrow: true })`${devslotBinary} reload test-slot`
  
  if (!result.ok) {
    fail(`Reload failed: ${result.stderr}`)
    return
  }
  
  // Verify slot still exists
  if (!await fs.pathExists('slots/test-slot/repo')) {
    fail('Slot worktree missing after reload')
    return
  }
  
  if (!result.stdout.includes('reloaded successfully')) {
    fail('Output missing success message')
    return
  }
  
  pass()
}

async function testReloadMissingWorktree() {
  await setupTest('reload_missing_worktree')
  
  const repo1 = await createTestRepo('repo1')
  const repo2 = await createTestRepo('repo2')
  
  // Start with one repo
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo1
    url: ${repo1}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create my-slot`
  
  // Add second repo to config
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo1
    url: ${repo1}
  - name: repo2
    url: ${repo2}
`)
  
  // Init the new repo
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Reload should create missing worktree
  const result = await $({ nothrow: true })`${devslotBinary} reload my-slot`
  
  if (!result.ok) {
    fail(`Reload failed: ${result.stderr}`)
    return
  }
  
  // Check both worktrees exist
  if (!await fs.pathExists('slots/my-slot/repo1')) {
    fail('Original worktree missing')
    return
  }
  
  if (!await fs.pathExists('slots/my-slot/repo2')) {
    fail('New worktree not created by reload')
    return
  }
  
  pass()
}

async function testReloadNonExistent() {
  await setupTest('reload_nonexistent')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories: []
`)
  
  const result = await $({ nothrow: true })`${devslotBinary} reload nonexistent-slot`
  
  if (result.ok) {
    fail('Reloading non-existent slot should fail')
    return
  }
  
  const errorOutput = result.stderr + result.stdout
  if (!errorOutput.includes('does not exist')) {
    fail(`Expected "does not exist" error, got: ${errorOutput}`)
    return
  }
  
  pass()
}

async function testReloadWithHook() {
  await setupTest('reload_with_hook')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create post-reload hook
  await $`mkdir -p hooks`
  await fs.writeFile('hooks/post-reload', `#!/bin/bash
echo "Post-reload hook executed for slot: $DEVSLOT_SLOT_NAME"
echo "$DEVSLOT_SLOT_NAME reloaded at $(date)" >> reload.log
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create hook-slot`
  
  // Reload with hook
  const result = await $({ nothrow: true })`${devslotBinary} reload hook-slot`
  
  if (!result.ok) {
    fail(`Reload with hook failed: ${result.stderr}`)
    return
  }
  
  // Verify hook was executed
  if (!await fs.pathExists('reload.log')) {
    fail('Post-reload hook was not executed')
    return
  }
  
  const logContent = await fs.readFile('reload.log', 'utf-8')
  if (!logContent.includes('hook-slot reloaded')) {
    fail('Hook did not receive correct slot name')
    return
  }
  
  pass()
}

async function testReloadFailingHook() {
  await setupTest('reload_failing_hook')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create failing post-reload hook
  await $`mkdir -p hooks`
  await fs.writeFile('hooks/post-reload', `#!/bin/bash
echo "Post-reload hook failing!"
exit 1
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create fail-slot`
  
  // Try to reload - should fail
  const result = await $({ nothrow: true })`${devslotBinary} reload fail-slot`
  
  if (result.ok) {
    fail('Reload should fail when post-reload hook fails')
    return
  }
  
  const errorOutput = result.stderr + result.stdout
  if (!errorOutput.includes('post-reload hook failed')) {
    fail('Error message should mention hook failure')
    return
  }
  
  // Slot should still exist
  if (!await fs.pathExists('slots/fail-slot')) {
    fail('Slot removed after failed reload')
    return
  }
  
  pass()
}

async function testReloadAfterRepoRemoval() {
  await setupTest('reload_after_repo_removal')
  
  const repo1 = await createTestRepo('keep-repo')
  const repo2 = await createTestRepo('remove-repo')
  
  // Start with two repos
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: keep
    url: ${repo1}
  - name: remove
    url: ${repo2}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create multi-slot`
  
  // Remove one repo from config
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: keep
    url: ${repo1}
`)
  
  // Reload should still work (just ignores the extra worktree)
  const result = await $({ nothrow: true })`${devslotBinary} reload multi-slot`
  
  if (!result.ok) {
    fail(`Reload after repo removal failed: ${result.stderr}`)
    return
  }
  
  // Original worktree should remain
  if (!await fs.pathExists('slots/multi-slot/keep')) {
    fail('Remaining repo worktree missing')
    return
  }
  
  // Removed repo worktree might still exist (reload doesn't remove)
  // This is expected behavior
  
  pass()
}

async function testReloadCorruptedWorktree() {
  await setupTest('reload_corrupted_worktree')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create corrupt-slot`
  
  // Corrupt the worktree by removing .git file
  await $`rm -f slots/corrupt-slot/repo/.git`
  
  // Reload should handle this gracefully
  const result = await $({ nothrow: true })`${devslotBinary} reload corrupt-slot`
  
  // The behavior here depends on implementation
  // It might fail or try to recreate the worktree
  if (!result.ok) {
    // If it fails, check for reasonable error message
    const errorOutput = result.stderr + result.stdout
    if (!errorOutput.includes('worktree') && !errorOutput.includes('git')) {
      fail('Error should mention worktree or git issue')
      return
    }
  }
  
  pass()
}

// Main test runner
async function runTests() {
  echo('Running devslot reload E2E tests...')
  echo('================================\n')
  
  // Check if binary exists
  if (!await fs.pathExists(devslotBinary)) {
    echo(chalk.red('Error: devslot binary not found. Run "make build" first.'))
    process.exit(1)
  }
  
  // Run all tests
  await testBasicReload()
  await testReloadMissingWorktree()
  await testReloadNonExistent()
  await testReloadWithHook()
  await testReloadFailingHook()
  await testReloadAfterRepoRemoval()
  await testReloadCorruptedWorktree()
  
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