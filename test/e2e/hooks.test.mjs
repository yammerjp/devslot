#!/usr/bin/env zx

// E2E tests for devslot hooks functionality

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
async function testPostCreateHook() {
  await setupTest('post_create_hook')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create post-create hook
  await $`mkdir -p hooks`
  await fs.writeFile('hooks/post-create', `#!/bin/bash
echo "🎉 Post-create hook executed!"
echo "Slot name: $DEVSLOT_SLOT_NAME"
echo "Slot directory: $DEVSLOT_SLOT_DIR"
echo "Project root: $DEVSLOT_ROOT"
echo "Repos directory: $DEVSLOT_REPOS_DIR"

# Create marker file
echo "Created at $(date)" > "$DEVSLOT_SLOT_DIR/hook-marker.txt"

# Log to project root
echo "$DEVSLOT_SLOT_NAME created" >> "$DEVSLOT_ROOT/created-slots.log"
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Create slot
  const result = await $({ nothrow: true })`${devslotBinary} create test-slot`
  
  if (!result.ok) {
    fail(`Create failed: ${result.stderr}`)
    return
  }
  
  // Verify hook was executed
  if (!await fs.pathExists('slots/test-slot/hook-marker.txt')) {
    fail('Post-create hook did not create marker file')
    return
  }
  
  if (!await fs.pathExists('created-slots.log')) {
    fail('Post-create hook did not create log file')
    return
  }
  
  const logContent = await fs.readFile('created-slots.log', 'utf-8')
  if (!logContent.includes('test-slot created')) {
    fail('Hook did not log correct slot name')
    return
  }
  
  // Check hook output in command output
  if (!result.stdout.includes('Post-create hook executed!')) {
    fail('Hook output not shown in command output')
    return
  }
  
  pass()
}

async function testPreDestroyHook() {
  await setupTest('pre_destroy_hook')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create pre-destroy hook
  await $`mkdir -p hooks`
  await fs.writeFile('hooks/pre-destroy', `#!/bin/bash
echo "🗑️ Pre-destroy hook executed!"
echo "About to destroy: $DEVSLOT_SLOT_NAME"

# Save slot contents list
ls -la "$DEVSLOT_SLOT_DIR" > "$DEVSLOT_ROOT/destroyed-$DEVSLOT_SLOT_NAME-contents.txt"

# Create backup marker
mkdir -p "$DEVSLOT_ROOT/backups"
echo "Destroyed at $(date)" > "$DEVSLOT_ROOT/backups/$DEVSLOT_SLOT_NAME.backup"
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create destroy-test`
  
  // Add some content to the slot
  await fs.writeFile('slots/destroy-test/test-file.txt', 'Important data')
  
  // Destroy slot
  const result = await $({ nothrow: true })`${devslotBinary} destroy destroy-test`
  
  if (!result.ok) {
    fail(`Destroy failed: ${result.stderr}`)
    return
  }
  
  // Verify hook was executed
  if (!await fs.pathExists('destroyed-destroy-test-contents.txt')) {
    fail('Pre-destroy hook did not create contents list')
    return
  }
  
  if (!await fs.pathExists('backups/destroy-test.backup')) {
    fail('Pre-destroy hook did not create backup marker')
    return
  }
  
  // Check hook output
  if (!result.stdout.includes('Pre-destroy hook executed!')) {
    fail('Hook output not shown')
    return
  }
  
  pass()
}

async function testPostReloadHook() {
  await setupTest('post_reload_hook')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create post-reload hook
  await $`mkdir -p hooks`
  await fs.writeFile('hooks/post-reload', `#!/bin/bash
echo "🔄 Post-reload hook executed!"
echo "Reloaded slot: $DEVSLOT_SLOT_NAME"

# Update timestamp
echo "Last reloaded at $(date)" > "$DEVSLOT_SLOT_DIR/last-reload.txt"

# Count reloads
RELOAD_COUNT=0
if [ -f "$DEVSLOT_ROOT/reload-count.txt" ]; then
  RELOAD_COUNT=$(cat "$DEVSLOT_ROOT/reload-count.txt")
fi
RELOAD_COUNT=$((RELOAD_COUNT + 1))
echo $RELOAD_COUNT > "$DEVSLOT_ROOT/reload-count.txt"
echo "Reload count: $RELOAD_COUNT"
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  await $({ nothrow: true })`${devslotBinary} create reload-test`
  
  // Reload multiple times
  for (let i = 0; i < 3; i++) {
    const result = await $({ nothrow: true })`${devslotBinary} reload reload-test`
    if (!result.ok) {
      fail(`Reload ${i+1} failed: ${result.stderr}`)
      return
    }
  }
  
  // Verify hook was executed
  if (!await fs.pathExists('slots/reload-test/last-reload.txt')) {
    fail('Post-reload hook did not create timestamp file')
    return
  }
  
  const reloadCount = await fs.readFile('reload-count.txt', 'utf-8')
  if (reloadCount.trim() !== '3') {
    fail(`Expected reload count 3, got ${reloadCount.trim()}`)
    return
  }
  
  pass()
}

async function testHookEnvironmentVariables() {
  await setupTest('hook_env_vars')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create hook that logs all environment variables
  await $`mkdir -p hooks`
  await fs.writeFile('hooks/post-create', `#!/bin/bash
echo "=== Environment Variables ==="
echo "DEVSLOT_ROOT=$DEVSLOT_ROOT"
echo "DEVSLOT_SLOT_NAME=$DEVSLOT_SLOT_NAME"
echo "DEVSLOT_SLOT_DIR=$DEVSLOT_SLOT_DIR"
echo "DEVSLOT_REPOS_DIR=$DEVSLOT_REPOS_DIR"

# Validate paths
if [ ! -d "$DEVSLOT_ROOT" ]; then
  echo "ERROR: DEVSLOT_ROOT does not exist"
  exit 1
fi

if [ "$DEVSLOT_SLOT_NAME" != "env-test" ]; then
  echo "ERROR: DEVSLOT_SLOT_NAME is incorrect"
  exit 1
fi

if [ ! -d "$DEVSLOT_SLOT_DIR" ]; then
  echo "ERROR: DEVSLOT_SLOT_DIR does not exist"
  exit 1
fi

if [ ! -d "$DEVSLOT_REPOS_DIR" ]; then
  echo "ERROR: DEVSLOT_REPOS_DIR does not exist"
  exit 1
fi

echo "All environment variables validated successfully!"
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  const result = await $({ nothrow: true })`${devslotBinary} create env-test`
  
  if (!result.ok) {
    fail(`Create failed: ${result.stderr}`)
    return
  }
  
  if (!result.stdout.includes('All environment variables validated successfully!')) {
    fail('Hook environment variables validation failed')
    return
  }
  
  pass()
}

async function testHookFailureHandling() {
  await setupTest('hook_failure_handling')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create failing hooks
  await $`mkdir -p hooks`
  
  // Post-create hook that fails
  await fs.writeFile('hooks/post-create', `#!/bin/bash
echo "Post-create hook starting..."
echo "ERROR: Simulated failure!"
exit 1
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Test post-create failure (should cleanup slot)
  const createResult = await $({ nothrow: true })`${devslotBinary} create fail-create`
  
  if (createResult.ok) {
    fail('Create should fail when post-create hook fails')
    return
  }
  
  // Check if slot was cleaned up
  // The slot directory might be created but should be empty or removed
  const slotPath = 'slots/fail-create'
  if (await fs.pathExists(slotPath)) {
    try {
      const contents = await fs.readdir(slotPath)
      // If directory exists, it should be empty (cleanup might leave empty dir)
      if (contents.length > 0) {
        fail(`Slot should be cleaned up after post-create hook failure, but contains: ${contents.join(', ')}`)
        return
      }
      // Empty directory is acceptable - cleanup was attempted
    } catch (e) {
      // Directory might have been removed between checks - this is fine
    }
  }
  
  // Create slot without hook for destroy test
  await $`rm hooks/post-create`
  
  // Create pre-destroy hook that fails
  await fs.writeFile('hooks/pre-destroy', `#!/bin/bash
echo "Pre-destroy hook blocking destruction!"
exit 1
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} create fail-destroy`
  
  // Test pre-destroy failure (should prevent destruction)
  const destroyResult = await $({ nothrow: true })`${devslotBinary} destroy fail-destroy`
  
  if (destroyResult.ok) {
    fail('Destroy should fail when pre-destroy hook fails')
    return
  }
  
  if (!await fs.pathExists('slots/fail-destroy')) {
    fail('Slot should not be destroyed when pre-destroy hook fails')
    return
  }
  
  pass()
}

async function testMissingHooks() {
  await setupTest('missing_hooks')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // No hooks directory or scripts
  await $({ nothrow: true })`${devslotBinary} init`
  
  // Commands should work without hooks
  const createResult = await $({ nothrow: true })`${devslotBinary} create no-hooks`
  if (!createResult.ok) {
    fail('Create should work without hooks')
    return
  }
  
  const reloadResult = await $({ nothrow: true })`${devslotBinary} reload no-hooks`
  if (!reloadResult.ok) {
    fail('Reload should work without hooks')
    return
  }
  
  const destroyResult = await $({ nothrow: true })`${devslotBinary} destroy no-hooks`
  if (!destroyResult.ok) {
    fail('Destroy should work without hooks')
    return
  }
  
  pass()
}

async function testHookWithComplexOperations() {
  await setupTest('hook_complex_operations')
  
  const repo = await createTestRepo('repo')
  await fs.writeFile('devslot.yaml', `version: 1
repositories:
  - name: repo
    url: ${repo}
`)
  
  // Create complex post-create hook
  await $`mkdir -p hooks`
  await fs.writeFile('hooks/post-create', `#!/bin/bash
echo "Running complex post-create operations..."

# Create directory structure
mkdir -p "$DEVSLOT_SLOT_DIR/.devcontainer"
mkdir -p "$DEVSLOT_SLOT_DIR/.vscode"

# Create devcontainer.json
cat > "$DEVSLOT_SLOT_DIR/.devcontainer/devcontainer.json" << 'EOF'
{
  "name": "DevSlot Environment",
  "workspaceFolder": "/workspace",
  "customizations": {
    "vscode": {
      "extensions": ["ms-vscode.cpptools"]
    }
  }
}
EOF

# Create VS Code settings
cat > "$DEVSLOT_SLOT_DIR/.vscode/settings.json" << 'EOF'
{
  "editor.formatOnSave": true,
  "editor.tabSize": 2
}
EOF

# Create a script in the worktree
cat > "$DEVSLOT_SLOT_DIR/repo/setup.sh" << 'EOF'
#!/bin/bash
echo "Development environment ready!"
EOF
chmod +x "$DEVSLOT_SLOT_DIR/repo/setup.sh"

echo "Complex setup completed!"
`, { mode: 0o755 })
  
  await $({ nothrow: true })`${devslotBinary} init`
  
  const result = await $({ nothrow: true })`${devslotBinary} create complex-slot`
  
  if (!result.ok) {
    fail(`Create with complex hook failed: ${result.stderr}`)
    return
  }
  
  // Verify complex operations
  if (!await fs.pathExists('slots/complex-slot/.devcontainer/devcontainer.json')) {
    fail('Hook did not create devcontainer.json')
    return
  }
  
  if (!await fs.pathExists('slots/complex-slot/.vscode/settings.json')) {
    fail('Hook did not create VS Code settings')
    return
  }
  
  if (!await fs.pathExists('slots/complex-slot/repo/setup.sh')) {
    fail('Hook did not create setup script in worktree')
    return
  }
  
  // Check if setup.sh is executable
  const stats = await fs.stat('slots/complex-slot/repo/setup.sh')
  if (!(stats.mode & 0o100)) {
    fail('Setup script is not executable')
    return
  }
  
  pass()
}

// Main test runner
async function runTests() {
  echo('Running devslot hooks E2E tests...')
  echo('================================\n')
  
  // Check if binary exists
  if (!await fs.pathExists(devslotBinary)) {
    echo(chalk.red('Error: devslot binary not found. Run "make build" first.'))
    process.exit(1)
  }
  
  // Run all tests
  await testPostCreateHook()
  await testPreDestroyHook()
  await testPostReloadHook()
  await testHookEnvironmentVariables()
  await testHookFailureHandling()
  await testMissingHooks()
  await testHookWithComplexOperations()
  
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