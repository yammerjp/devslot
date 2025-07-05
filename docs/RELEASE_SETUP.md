# リリースセットアップガイド

このプロジェクトではtagprを使用して自動リリース管理を行います。3つのセットアップ方法があります。

## オプション1: 基本セットアップ（手動タグ付け）

最もシンプルな設定で、デフォルトの`GITHUB_TOKEN`を使用します。ただし、tagprがタグを作成してもリリースワークフローは**自動的にトリガーされません**。

**制限事項**: tagprがタグを作成した後、手動でリリースワークフローを実行する必要があります。

## オプション2: Personal Access Token (PAT)

PATを使用すると、tagprがリリースワークフローを自動的にトリガーできます。

### セットアップ手順:

1. Personal Access Tokenを作成:
   - GitHub Settings > Developer settings > Personal access tokens へ移動
   - `repo`と`workflow`スコープを持つ新しいトークンを作成
   - トークンをコピー

2. リポジトリのシークレットに追加:
   - リポジトリの Settings > Secrets and variables > Actions へ移動
   - `TAGPR_PAT`という名前で新しいシークレットを追加
   - コピーしたPATを値として貼り付け

3. ワークフローは利用可能な場合、自動的にPATを使用します

## オプション3: GitHub App（組織での使用推奨）

GitHub AppsはPATと比較して、より優れたセキュリティと管理機能を提供します。

### セットアップ手順:

1. GitHub Appを作成:
   - Settings > Developer settings > GitHub Apps へ移動
   - 「New GitHub App」をクリック
   - 必要項目を入力:
     - Name: `<your-org>-tagpr-bot` （または任意の名前）
     - Homepage URL: リポジトリのURL
     - Webhook: 「Active」のチェックを外す
   - 権限設定:
     - Repository permissions:
       - Contents: Read & Write
       - Metadata: Read
       - Pull requests: Read & Write
       - Issues: Read & Write
       - Actions: Read
     - Account permissions: なし
   - Where can this GitHub App be installed: 「Only on this account」
   - 「Create GitHub App」をクリック

2. 秘密鍵を生成して保存:
   - 作成したAppの設定で「Private keys」までスクロール
   - 「Generate a private key」をクリック
   - ダウンロードされた.pemファイルを保存

3. Appをインストール:
   - App設定で「Install App」をクリック
   - 対象のリポジトリを選択

4. リポジトリを設定:
   - App ID（App設定に表示）をメモ
   - リポジトリの Settings > Secrets and variables > Actions へ移動
   - シークレット`APP_PRIVATE_KEY`に.pemファイルの内容を追加
   - 変数`APP_ID`にApp IDを追加

5. サンプルワークフローを使用:
   ```bash
   cp .github/workflows/tagpr-with-app.yaml.example .github/workflows/tagpr.yaml
   ```

## 比較表

| 機能 | 基本 | PAT | GitHub App |
|------|------|-----|------------|
| リリース自動トリガー | ❌ | ✅ | ✅ |
| セキュリティ | ✅ | ⚠️ | ✅ |
| ユーザー非依存 | ✅ | ❌ | ✅ |
| 有効期限 | なし | 任意 | 1時間（自動更新） |
| 監査証跡 | 基本 | ユーザーベース | Appベース |
| セットアップの複雑さ | 低 | 中 | 高 |

## トラブルシューティング

### リリースワークフローがトリガーされない

1. PATまたはGitHub Appが正しい権限を持っているか確認
2. シークレット/変数名がワークフローと一致しているか確認
3. Actionsログで権限エラーがないか確認

### 権限エラー

以下の設定が有効になっているか確認:
Settings > Actions > General > Workflow permissions > 「Allow GitHub Actions to create and approve pull requests」