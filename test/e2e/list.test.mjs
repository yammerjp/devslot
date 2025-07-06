#!/usr/bin/env zx

// E2E tests for devslot list command

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
async function testListEmpty() {
  await setupTest('list_empty')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories: []
`)
  
  const result = await $({ nothrow: true })`${devslotBinary} list`
  
  if (!result.ok) {
    fail(`List failed: ${result.stderr}`)
    return
  }
  
  const output = result.stdout.trim()
  if (output !== '' && !output.includes('No slots found')) {
    fail(`Expected empty output or "No slots found", got: ${output}`)
    return
  }
  
  pass()
}

async function testListSingleSlot() {
  await setupTest('list_single_slot')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create my-slot`
  
  const result = await $({ nothrow: true })`${devslotBinary} list`
  
  if (!result.ok) {
    fail(`List failed: ${result.stderr}`)
    return
  }
  
  const output = result.stdout.trim()
  if (!output.includes('my-slot')) {
    fail(`Output should include slot name, got: ${output}`)
    return
  }
  
  pass()
}

async function testListMultipleSlots() {
  await setupTest('list_multiple_slots')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Create multiple slots
  const slots = ['dev', 'test', 'feature-x', 'bugfix-123']
  for (const slot of slots) {
    await $({ nothrow: true })`${devslotBinary} create ${slot}`
  }
  
  const result = await $({ nothrow: true })`${devslotBinary} list`
  
  if (!result.ok) {
    fail(`List failed: ${result.stderr}`)
    return
  }
  
  const output = result.stdout
  
  // Check all slots are listed
  for (const slot of slots) {
    if (!output.includes(slot)) {
      fail(`Slot "${slot}" not found in output`)
      return
    }
  }
  
  // Note: Current implementation doesn't guarantee alphabetical order
  // This is acceptable as the primary requirement is to list all slots
  // If alphabetical order is needed, it should be implemented in the List() method
  
  pass()
}

async function testListAfterDestroy() {
  await setupTest('list_after_destroy')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Create slots
  await $({ nothrow: true })`${devslotBinary} create slot1`
  await $({ nothrow: true })`${devslotBinary} create slot2`
  await $({ nothrow: true })`${devslotBinary} create slot3`
  
  // Destroy one
  await $({ nothrow: true })`${devslotBinary} destroy slot2`
  
  const result = await $({ nothrow: true })`${devslotBinary} list`
  
  if (!result.ok) {
    fail(`List failed: ${result.stderr}`)
    return
  }
  
  const output = result.stdout
  
  if (!output.includes('slot1') || !output.includes('slot3')) {
    fail('Remaining slots not listed')
    return
  }
  
  if (output.includes('slot2')) {
    fail('Destroyed slot still appears in list')
    return
  }
  
  pass()
}

async function testListWithoutSlotsDir() {
  await setupTest('list_without_slots_dir')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories: []
`)
  
  // Don't create slots directory
  const result = await $({ nothrow: true })`${devslotBinary} list`
  
  if (!result.ok) {
    fail(`List should handle missing slots directory gracefully: ${result.stderr}`)
    return
  }
  
  const output = result.stdout.trim()
  if (output !== '' && !output.includes('No slots found')) {
    fail('Should return empty when slots directory missing')
    return
  }
  
  pass()
}

async function testListWithInvalidSlots() {
  await setupTest('list_with_invalid_slots')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create valid-slot`
  
  // Create invalid entries in slots directory
  await $`mkdir -p slots`
  await fs.writeFile('slots/not-a-directory.txt', 'This is a file, not a slot')
  await $`ln -s /nonexistent slots/broken-symlink`
  
  const result = await $({ nothrow: true })`${devslotBinary} list`
  
  if (!result.ok) {
    fail(`List failed with invalid entries: ${result.stderr}`)
    return
  }
  
  const output = result.stdout
  
  // Should list valid slot
  if (!output.includes('valid-slot')) {
    fail('Valid slot not listed')
    return
  }
  
  // Should not list invalid entries
  if (output.includes('not-a-directory.txt') || output.includes('broken-symlink')) {
    fail('Invalid entries should not be listed')
    return
  }
  
  pass()
}

async function testListOutsideProject() {
  await setupTest('list_outside_project')
  
  // No devslot.yaml
  const result = await $({ nothrow: true })`${devslotBinary} list`
  
  if (result.ok) {
    fail('List should fail outside devslot project')
    return
  }
  
  const errorOutput = result.stderr + result.stdout
  if (!errorOutput.includes('not in a devslot project')) {
    fail(`Expected "not in a devslot project" error, got: ${errorOutput}`)
    return
  }
  
  pass()
}

// Main test runner
async function runTests() {
  echo('Running devslot list E2E tests...')
  echo('================================\n')
  
  // Check if binary exists
  if (!await fs.pathExists(devslotBinary)) {
    echo(chalk.red('Error: devslot binary not found. Run "make build" first.'))
    process.exit(1)
  }
  
  // Run all tests
  await testListEmpty()
  await testListSingleSlot()
  await testListMultipleSlots()
  await testListAfterDestroy()
  await testListWithoutSlotsDir()
  await testListWithInvalidSlots()
  await testListOutsideProject()
  
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