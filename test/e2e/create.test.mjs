#!/usr/bin/env zx

// E2E tests for devslot create command

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
    // Remove .git directory if it exists
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
async function testBasicCreate() {
  await setupTest('basic_create')
  
  // Create test repositories
  const repo1 = await createTestRepo('repo1')
  const repo2 = await createTestRepo('repo2')
  
  // Create devslot.yaml
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo1
    url: ${repo1}
  - name: repo2
    url: ${repo2}
`)
  
  // Initialize repositories
  const initResult = await $({ nothrow: true })`${devslotBinary} init`
  if (!initResult.ok) {
    fail(`Init failed: ${initResult.stderr}`)
    return
  }
  
  // Create a slot
  const result = await $({ nothrow: true })`${devslotBinary} create test-slot`
  
  // Assertions
  if (!result.ok) {
    fail(`Create failed: ${result.stderr}`)
    return
  }
  
  if (!await fs.pathExists('slots/test-slot')) {
    fail('slots/test-slot does not exist')
    return
  }
  
  if (!await fs.pathExists('slots/test-slot/repo1')) {
    fail('slots/test-slot/repo1 does not exist')
    return
  }
  
  if (!await fs.pathExists('slots/test-slot/repo2')) {
    fail('slots/test-slot/repo2 does not exist')
    return
  }
  
  // Check if they are valid git repositories
  const gitCheck = await $({ nothrow: true })`cd slots/test-slot/repo1 && git status`
  if (!gitCheck.ok) {
    fail('repo1 is not a valid git repository')
    return
  }
  
  if (!result.stdout.includes('created successfully')) {
    fail('Output missing success message')
    return
  }
  
  pass()
}

async function testCreateWithBranch() {
  await setupTest('create_with_branch')
  
  const repo = await createTestRepo('repo-with-branches')
  
  // Create a feature branch in the test repo
  const oldCwd = process.cwd()
  cd(repo)
  try {
    await $`git checkout -b feature-branch`
    await fs.writeFile('feature.txt', 'Feature content')
    await $`git add feature.txt`
    await $`git commit -m "Add feature"`
    await $`git checkout master || git checkout main`
  } finally {
    cd(oldCwd)
  }
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Initialize
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Create slot with specific branch
  const result = await $({ nothrow: true })`${devslotBinary} create -b feature-branch feature-slot`
  
  if (!result.ok) {
    fail(`Create with branch failed: ${result.stderr}`)
    return
  }
  
  // Check if on correct branch
  const branchCheck = await $({ nothrow: true })`cd slots/feature-slot/repo && git branch --show-current`
  if (!branchCheck.ok || branchCheck.stdout.trim() !== 'feature-branch') {
    fail('Not on expected branch')
    return
  }
  
  // Check if feature file exists
  if (!await fs.pathExists('slots/feature-slot/repo/feature.txt')) {
    fail('Feature file not found in worktree')
    return
  }
  
  pass()
}

async function testDuplicateSlot() {
  await setupTest('duplicate_slot')
  
  const repo = await createTestRepo('repo')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Create first slot
  const firstResult = await $({ nothrow: true })`${devslotBinary} create my-slot`
  if (!firstResult.ok) {
    fail(`First create failed: ${firstResult.stderr}`)
    return
  }
  
  // Try to create duplicate
  const result = await $({ nothrow: true })`${devslotBinary} create my-slot`
  
  if (result.ok) {
    fail('Creating duplicate slot should have failed')
    return
  }
  
  const errorOutput = result.stderr + result.stdout
  if (!errorOutput.includes('already exists')) {
    fail(`Expected "already exists" error, got: ${errorOutput}`)
    return
  }
  
  pass()
}

async function testInvalidSlotName() {
  await setupTest('invalid_slot_name')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories: []
`)
  
  // Try invalid names
  const invalidNames = ['slot/with/slash', 'slot\\with\\backslash', '..', '.']
  
  for (const name of invalidNames) {
    const result = await $({ nothrow: true })`${devslotBinary} create "${name}"`
    
    if (result.ok) {
      fail(`Creating slot with invalid name "${name}" should have failed`)
      return
    }
  }
  
  pass()
}

async function testCreateWithoutInit() {
  await setupTest('create_without_init')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: https://example.com/repo.git
`)
  
  // Try to create slot without init
  const result = await $({ nothrow: true })`${devslotBinary} create test-slot`
  
  if (result.ok) {
    fail('Create should have failed without init')
    return
  }
  
  const errorOutput = result.stderr + result.stdout
  if (!errorOutput.includes('does not exist')) {
    fail(`Expected "does not exist" error, got: ${errorOutput}`)
    return
  }
  
  pass()
}

async function testMultipleSlots() {
  await setupTest('multiple_slots')
  
  const repo = await createTestRepo('shared-repo')
  
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Create multiple slots with different branches
  const slots = [
    { name: 'dev', branch: 'dev-branch' },
    { name: 'test', branch: 'test-branch' },
    { name: 'feature', branch: 'feature-branch' }
  ]
  
  for (const slot of slots) {
    const result = await $({ nothrow: true })`${devslotBinary} create -b ${slot.branch} ${slot.name}`
    if (!result.ok) {
      fail(`Failed to create slot ${slot.name}: ${result.stderr}`)
      return
    }
  }
  
  // Verify all slots exist
  for (const slot of slots) {
    if (!await fs.pathExists(`slots/${slot.name}/repo`)) {
      fail(`Slot ${slot.name} not created properly`)
      return
    }
  }
  
  // Verify they are independent worktrees
  // Create different files in each
  for (const slot of slots) {
    await fs.writeFile(`slots/${slot.name}/repo/file-${slot.name}.txt`, `Content for ${slot.name}`)
  }
  
  // Check that files don't exist in other worktrees
  for (let i = 0; i < slots.length; i++) {
    for (let j = 0; j < slots.length; j++) {
      if (i !== j) {
        if (await fs.pathExists(`slots/${slots[i].name}/repo/file-${slots[j].name}.txt`)) {
          fail('Worktrees are not independent')
          return
        }
      }
    }
  }
  
  pass()
}

// Main test runner
async function runTests() {
  echo('Running devslot create E2E tests...')
  echo('================================\n')
  
  // Check if binary exists
  if (!await fs.pathExists(devslotBinary)) {
    echo(chalk.red('Error: devslot binary not found. Run "make build" first.'))
    process.exit(1)
  }
  
  // Run all tests
  await testBasicCreate()
  await testCreateWithBranch()
  await testDuplicateSlot()
  await testInvalidSlotName()
  await testCreateWithoutInit()
  await testMultipleSlots()
  
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