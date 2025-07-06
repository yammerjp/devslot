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
  
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate .`
  
  if (!result.ok) {
    fail(`Boilerplate failed: ${result.stderr}`)
    return
  }
  
  // Check created directories
  const dirs = ['hooks', 'repos', 'slots']
  for (const dir of dirs) {
    if (!await fs.pathExists(dir)) {
      fail(`Directory ${dir} was not created`)
      return
    }
  }
  
  // Check created files
  const files = [
    'devslot.yaml',
    '.gitignore',
    'hooks/post-create',
    'hooks/pre-destroy',
    'hooks/post-reload'
  ]
  
  for (const file of files) {
    if (!await fs.pathExists(file)) {
      fail(`File ${file} was not created`)
      return
    }
  }
  
  // Check hook permissions
  const hooks = ['post-create', 'pre-destroy', 'post-reload']
  for (const hook of hooks) {
    const stat = await fs.stat(`hooks/${hook}`)
    if (!(stat.mode & 0o100)) {
      fail(`Hook ${hook} is not executable`)
      return
    }
  }
  
  // Check devslot.yaml content
  const yamlContent = await fs.readFile('devslot.yaml', 'utf-8')
  if (!yamlContent.includes('version: 1')) {
    fail('devslot.yaml missing version')
    return
  }
  
  pass()
}

async function testBoilerplateNewDir() {
  await setupTest('boilerplate_new_dir')
  
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate my-new-project`
  
  if (!result.ok) {
    fail(`Boilerplate failed: ${result.stderr}`)
    return
  }
  
  // Check that directory was created
  if (!await fs.pathExists('my-new-project')) {
    fail('Project directory was not created')
    return
  }
  
  // Check files in the new directory
  if (!await fs.pathExists('my-new-project/devslot.yaml')) {
    fail('devslot.yaml was not created in new directory')
    return
  }
  
  if (!await fs.pathExists('my-new-project/hooks/post-create')) {
    fail('Hooks were not created in new directory')
    return
  }
  
  pass()
}

async function testBoilerplateExistingFiles() {
  await setupTest('boilerplate_existing_files')
  
  // Create existing .gitignore
  await fs.writeFile('.gitignore', 'node_modules/\n*.log\n')
  
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate .`
  
  if (!result.ok) {
    fail(`Boilerplate failed: ${result.stderr}`)
    return
  }
  
  // Check that .gitignore was updated
  const gitignore = await fs.readFile('.gitignore', 'utf-8')
  if (!gitignore.includes('node_modules/')) {
    fail('.gitignore lost existing content')
    return
  }
  
  if (!gitignore.includes('/repos/')) {
    fail('.gitignore missing devslot entries')
    return
  }
  
  pass()
}

async function testBoilerplateAbsolutePath() {
  await setupTest('boilerplate_absolute_path')
  
  const absolutePath = path.join(testDir, 'absolute-project')
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate ${absolutePath}`
  
  if (!result.ok) {
    fail(`Boilerplate failed: ${result.stderr}`)
    return
  }
  
  // Check that directory was created at absolute path
  if (!await fs.pathExists(absolutePath)) {
    fail('Project directory was not created at absolute path')
    return
  }
  
  if (!await fs.pathExists(path.join(absolutePath, 'devslot.yaml'))) {
    fail('devslot.yaml was not created at absolute path')
    return
  }
  
  pass()
}

async function testBoilerplateNoArgs() {
  await setupTest('boilerplate_no_args')
  
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate`
  
  if (result.ok) {
    fail('Boilerplate should fail without directory argument')
    return
  }
  
  // Check error message
  const errorOutput = result.stderr + result.stdout
  if (!errorOutput.includes('expected "<dir>"') && !errorOutput.includes('required')) {
    fail(`Unexpected error message: ${errorOutput}`)
    return
  }
  
  pass()
}

async function testBoilerplateHelp() {
  await setupTest('boilerplate_help')
  
  const result = await $({ nothrow: true })`${devslotBinary} boilerplate --help`
  
  if (!result.ok) {
    fail(`Help failed: ${result.stderr}`)
    return
  }
  
  const output = result.stdout
  
  // Check that help mentions directory argument
  if (!output.includes('<dir>') && !output.includes('Directory')) {
    fail('Help does not mention directory argument')
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
  await testBoilerplateNewDir()
  await testBoilerplateExistingFiles()
  await testBoilerplateAbsolutePath()
  await testBoilerplateNoArgs()
  await testBoilerplateHelp()
  
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