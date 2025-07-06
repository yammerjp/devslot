# devslot TODO

**重要**: このプロジェクトの開発は、Kent Beckが提唱しt-wadaが推奨する厳密なTDD（Red-Green-Refactor）により、堅牢なソフトウェアをインクリメンタルに構築していくこと。

## 現在の実装状況

### 完了済み
- ✅ Go moduleの初期化とkongのインストール
- ✅ 基本的なCLI構造とhelpコマンドの実装
- ✅ versionコマンドの実装 (-v/--version フラグ対応)
- ✅ boilerplateコマンドの実装
- ✅ 引数なしでhelpを表示する機能
- ✅ initコマンド（--allow-delete対応）
- ✅ createコマンド（-b/--branch対応）
- ✅ destroyコマンド
- ✅ reloadコマンド
- ✅ listコマンド
- ✅ doctorコマンド
- ✅ ロック機構（.devslot.lock）
- ✅ 設定ファイル管理（devslot.yaml）
- ✅ プロジェクトルート検出
- ✅ Git worktree操作
- ✅ フック機構（post-create, pre-destroy, post-reload）

### READMEとの差異・未実装機能
- ❌ boilerplateコマンドが`<dir>`引数を受け取らない（現在はカレントディレクトリ固定）
- ❌ READMEに記載の環境変数が実装と異なる
  - 実装: `DEVSLOT_SLOT`, `DEVSLOT_PROJECT_ROOT`
  - README: `DEVSLOT_ROOT`, `DEVSLOT_SLOT_NAME`, `DEVSLOT_SLOT_DIR`, `DEVSLOT_REPOS_DIR`
- ❌ devslot.yamlの簡易フォーマット（URLのみのリスト）非対応
  - READMEの例: `- https://github.com/example/app1`
  - 実装: `- name: app1.git` と `url: https://...` が必要
- ❌ createコマンドの-b/--branchオプションがREADMEに記載されていない
- ❌ フックファイルが必須と記載されているが、実際はオプション（doctorでは"optional"表示）

## 今後の改善項目

### 高優先度（本セッションで対応予定）
1. **環境変数の統一**
   - フックに渡される環境変数をREADMEの記載に合わせる
   - 実装変更が必要:
     - `DEVSLOT_PROJECT_ROOT` → `DEVSLOT_ROOT`
     - `DEVSLOT_SLOT` → `DEVSLOT_SLOT_NAME`
     - `DEVSLOT_SLOT_DIR`を追加（スロットディレクトリのフルパス）
     - `DEVSLOT_REPOS_DIR`を追加（reposディレクトリのフルパス）

2. **boilerplateコマンドの改善**
   - ディレクトリ引数のサポート追加（必須、`.`も可）
   - 指定されたディレクトリがない場合は作成
   - READMEに記載の通り `devslot boilerplate <dir>` 形式に

3. **READMEの修正**
   - devslot.yamlの簡易フォーマット（URLのみの配列）の記載を削除
   - 現在の実装に合わせて name/url 形式のみ記載
   - createコマンドの-b/--branchオプションの説明を追加

### 中優先度
4. **ドキュメント更新**
   - createコマンドの-b/--branchオプションをREADMEに追加
   - フックファイルの必須/オプション記載を実装に合わせる

5. **エラーメッセージの改善**
   - より分かりやすいエラーメッセージ
   - 解決方法の提示

6. **進捗表示の改善**
   - 大量リポジトリのclone時のプログレスバー
   - 並列処理の検討

### 低優先度
7. **診断機能の強化**
   - doctorコマンドでの自動修復機能
   - より詳細な整合性チェック

8. **パフォーマンス最適化**
   - 並列処理によるinit/createの高速化
   - キャッシュ機構の検討

## 技術的な考慮事項

### 現在使用している技術
- **Kong** - CLI フレームワーク
- **goccy/go-yaml** - YAML処理
- **標準ライブラリ** - ファイルロック、Git操作など

### テスト戦略
- ✅ 各コマンドに対するユニットテスト（TDD）
- ✅ E2Eテスト（zxベース）
- ✅ testutilパッケージによるテストヘルパー
- ✅ CI/CDでの自動テスト実行

### 今後の技術的改善
- プログレスバー表示（大量リポジトリ対応）
- 並列処理によるパフォーマンス向上
- より詳細なロギング機能

## コード品質の維持

- ✅ golangci-lintによる静的解析
- ✅ gofmtによるコードフォーマット
- ✅ go vetによるチェック
- ✅ テストカバレッジの計測
- ✅ GitHub Actionsによる自動CI/CD

