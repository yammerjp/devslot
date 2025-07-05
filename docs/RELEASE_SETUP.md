# Release Setup Guide

This project uses tagpr for automated release management. There are three ways to set it up:

## Option 1: Basic Setup (Manual Tagging)

The simplest setup uses the default `GITHUB_TOKEN`. However, this **will not** automatically trigger the release workflow when tagpr creates a tag.

**Limitation**: You'll need to manually trigger the release workflow after tagpr creates a tag.

## Option 2: Personal Access Token (PAT)

Using a PAT allows tagpr to trigger the release workflow automatically.

### Setup Steps:

1. Create a Personal Access Token:
   - Go to GitHub Settings > Developer settings > Personal access tokens
   - Create a new token with `repo` and `workflow` scopes
   - Copy the token

2. Add the token to repository secrets:
   - Go to repository Settings > Secrets and variables > Actions
   - Add a new secret named `TAGPR_PAT`
   - Paste your PAT as the value

3. The workflow will automatically use the PAT if available

## Option 3: GitHub App (Recommended for Organizations)

GitHub Apps provide better security and management compared to PATs.

### Setup Steps:

1. Create a GitHub App:
   - Go to Settings > Developer settings > GitHub Apps
   - Click "New GitHub App"
   - Fill in the required fields:
     - Name: `<your-org>-tagpr-bot` (or similar)
     - Homepage URL: Your repository URL
     - Webhook: Uncheck "Active"
   - Permissions:
     - Repository permissions:
       - Contents: Read & Write
       - Metadata: Read
       - Pull requests: Read & Write
       - Issues: Read & Write
       - Actions: Read
     - Account permissions: None
   - Where can this GitHub App be installed: "Only on this account"
   - Click "Create GitHub App"

2. Generate and store private key:
   - In your new App's settings, scroll to "Private keys"
   - Click "Generate a private key"
   - Save the downloaded .pem file

3. Install the App:
   - In the App settings, click "Install App"
   - Select your repository

4. Configure repository:
   - Note your App ID (shown in the App settings)
   - Go to repository Settings > Secrets and variables > Actions
   - Add secret `APP_PRIVATE_KEY` with the contents of the .pem file
   - Add variable `APP_ID` with your App ID

5. Use the example workflow:
   ```bash
   cp .github/workflows/tagpr-with-app.yaml.example .github/workflows/tagpr.yaml
   ```

## Comparison

| Feature | Basic | PAT | GitHub App |
|---------|-------|-----|------------|
| Auto-trigger release | ❌ | ✅ | ✅ |
| Security | ✅ | ⚠️ | ✅ |
| User-independent | ✅ | ❌ | ✅ |
| Expiration | N/A | Optional | 1 hour (auto-renewed) |
| Audit trail | Basic | User-based | App-based |
| Setup complexity | Low | Medium | High |

## Troubleshooting

### Release workflow not triggering

1. Check that the PAT or GitHub App has the correct permissions
2. Verify the secret/variable names match the workflow
3. Check the Actions log for any permission errors

### Permission errors

Ensure "Allow GitHub Actions to create and approve pull requests" is enabled in:
Settings > Actions > General > Workflow permissions