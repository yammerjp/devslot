#!/usr/bin/env zx

// E2E tests for devslot boilerplate command

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

// Tests
async function testBasicBoilerplate() {
  await setupTest('basic_boilerplate')
  
  // Run boilerplate in current directory
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate .`
  
  // Assertions
  if (!result.ok) {
    fail(`Boilerplate failed: ${result.stderr}`)
    return
  }
  
  // Check directories were created
  const expectedDirs = ['hooks', 'repos', 'slots']
  for (const dir of expectedDirs) {
    if (!await fs.pathExists(dir)) {
      fail(`Directory ${dir} was not created`)
      return
    }
  }
  
  // Check files were created
  if (!await fs.pathExists('devslot.yaml')) {
    fail('devslot.yaml was not created')
    return
  }
  
  if (!await fs.pathExists('.gitignore')) {
    fail('.gitignore was not created')
    return
  }
  
  // Check hook scripts were created with executable permissions
  const hookScripts = ['post-create', 'pre-destroy', 'post-reload']
  for (const hook of hookScripts) {
    const hookPath = path.join('hooks', hook)
    if (!await fs.pathExists(hookPath)) {
      fail(`Hook script ${hook} was not created`)
      return
    }
    
    // Check executable permissions
    const stats = await fs.stat(hookPath)
    const mode = stats.mode & parseInt('777', 8)
    if (mode !== parseInt('755', 8)) {
      fail(`Hook script ${hook} has wrong permissions: ${mode.toString(8)} (expected 755)`)
      return
    }
  }
  
  // Check devslot.yaml content
  const yamlContent = await fs.readFile('devslot.yaml', 'utf-8')
  if (!yamlContent.includes('version: 1')) {
    fail('devslot.yaml missing version')
    return
  }
  
  // Check .gitignore content
  const gitignoreContent = await fs.readFile('.gitignore', 'utf-8')
  if (!gitignoreContent.includes('/repos/') || !gitignoreContent.includes('/slots/')) {
    fail('.gitignore missing devslot directories')
    return
  }
  
  pass()
}

async function testBoilerplateInNewDir() {
  await setupTest('boilerplate_new_dir')
  
  // Run boilerplate to create new directory
  const newProjectDir = 'my-new-project'
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate ${newProjectDir}`
  
  if (!result.ok) {
    fail(`Boilerplate failed: ${result.stderr}`)
    return
  }
  
  // Check if directory was created
  if (!await fs.pathExists(newProjectDir)) {
    fail('New directory was not created')
    return
  }
  
  // Check structure inside new directory
  const expectedFiles = [
    'devslot.yaml',
    '.gitignore',
    'hooks/post-create',
    'hooks/pre-destroy',
    'hooks/post-reload'
  ]
  
  for (const file of expectedFiles) {
    const filePath = path.join(newProjectDir, file)
    if (!await fs.pathExists(filePath)) {
      fail(`File ${file} was not created in new directory`)
      return
    }
  }
  
  pass()
}

async function testBoilerplateExistingFiles() {
  await setupTest('boilerplate_existing_files')
  
  // Create existing .gitignore with custom content
  const existingGitignore = 'node_modules/\n*.log\ncustom-ignore/\n'
  await fs.writeFile('.gitignore', existingGitignore)
  
  // Create existing hook script
  await $`mkdir -p hooks`
  const existingHook = '#!/bin/bash\necho "Existing hook"\n'
  await fs.writeFile('hooks/post-create', existingHook, { mode: 0o755 })
  
  // Run boilerplate
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate .`
  
  if (!result.ok) {
    fail(`Boilerplate failed: ${result.stderr}`)
    return
  }
  
  // Check .gitignore was updated (not replaced)
  const gitignoreContent = await fs.readFile('.gitignore', 'utf-8')
  if (!gitignoreContent.includes('node_modules/')) {
    fail('.gitignore lost existing content')
    return
  }
  if (!gitignoreContent.includes('/repos/') || !gitignoreContent.includes('/slots/')) {
    fail('.gitignore missing devslot directories')
    return
  }
  
  // Check existing hook was preserved
  const hookContent = await fs.readFile('hooks/post-create', 'utf-8')
  if (hookContent !== existingHook) {
    fail('Existing hook was overwritten')
    return
  }
  
  pass()
}

async function testBoilerplateHookContent() {
  await setupTest('boilerplate_hook_content')
  
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate .`
  
  if (!result.ok) {
    fail(`Boilerplate failed: ${result.stderr}`)
    return
  }
  
  // Check hook content includes proper environment variables
  const postCreateContent = await fs.readFile('hooks/post-create', 'utf-8')
  const expectedVars = ['DEVSLOT_ROOT', 'DEVSLOT_SLOT_NAME', 'DEVSLOT_SLOT_DIR', 'DEVSLOT_REPOS_DIR']
  
  for (const envVar of expectedVars) {
    if (!postCreateContent.includes(envVar)) {
      fail(`post-create hook missing ${envVar} environment variable`)
      return
    }
  }
  
  // Check shebang
  if (!postCreateContent.startsWith('#!/bin/bash')) {
    fail('Hook script missing proper shebang')
    return
  }
  
  // Check for example code (should be commented out)
  if (!postCreateContent.includes('# echo')) {
    fail('Hook script missing example code')
    return
  }
  
  pass()
}

async function testBoilerplateAbsolutePath() {
  await setupTest('boilerplate_absolute_path')
  
  // Create absolute path
  const targetDir = path.join(testDir, 'absolute-path-test')
  await $`mkdir -p ${targetDir}`
  
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate ${targetDir}`
  
  if (!result.ok) {
    fail(`Boilerplate with absolute path failed: ${result.stderr}`)
    return
  }
  
  // Check files in target directory
  if (!await fs.pathExists(path.join(targetDir, 'devslot.yaml'))) {
    fail('devslot.yaml not created in absolute path')
    return
  }
  
  pass()
}

async function testBoilerplateInvalidPath() {
  await setupTest('boilerplate_invalid_path')
  
  // Try to create in a file (not directory)
  await fs.writeFile('notadir', 'content')
  
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate notadir/subdir`
  
  if (result.ok) {
    fail('Boilerplate should fail with invalid path')
    return
  }
  
  pass()
}

// Main test runner
async function runTests() {
  echo('Running devslot boilerplate E2E tests...')
  echo('================================\n')
  
  // Check if binary exists
  if (!await fs.pathExists(devslotBinary)) {
    echo(chalk.red('Error: devslot binary not found. Run "make build" first.'))
    process.exit(1)
  }
  
  // Run all tests
  await testBasicBoilerplate()
  await testBoilerplateInNewDir()
  await testBoilerplateExistingFiles()
  await testBoilerplateHookContent()
  await testBoilerplateAbsolutePath()
  await testBoilerplateInvalidPath()
  
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