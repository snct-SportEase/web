# バグ報告 / 既知の問題 (2026年1月20日更新)

現在確認されているバグおよび課題の一覧です。

## 未解決のバグ

### 1. テストコードのコンパイルエラー (Backend)
**重要度**: 高
**場所**: `backapp/tests/handler`
**説明**: バックエンドのテストを実行した際、モックオブジェクトがインターフェースを満たしていないためコンパイルエラーが発生します。
- `MockNoonGameRepository` が `repository.NoonGameRepository` インターフェースに追加されたメソッド（`GetTemplateDefaultGroups`, `SaveTemplateDefaultGroups`）を実装していません。
- **再現方法**: `backapp` ディレクトリで `go test ./...` を実行。

### 2. イベントIDのハードコード (Frontend)
**重要度**: 中
**場所**: `frontapp/src/routes/dashboard/admin/vorting-mvp/+page.svelte`
**説明**: MVP投票画面において、イベントIDが `1` に固定（ハードコード）されています。
- **影響**: IDが1以外のイベントでこの機能を使用すると正しく動作しない可能性があります。
- **修正**: アクティブなイベントIDをAPIから動的に取得するように修正する必要があります。 (`TODO: Get the active event id` というコメントが残っています)

### 3. 初期セットアップの状態
**重要度**: 低 (環境構築)
**場所**: プロジェクト全体
**説明**: プロジェクトの初期状態として、いくつかのファイルやディレクトリが不足しています。
- `frontapp/node_modules`: `npm install` が実行されていません。
- `.env`: `backapp` と `frontapp` の両方で `.env.sample` から `.env` を作成し、環境変数を設定する必要があります。

---

## 修正済みのバグ (2026/01/20 対応)

以下のバグは修正が適用されました。

1.  **環境変数の読み込みパス (Backend)**
    - 以前は `../../.env` を参照していましたが、カレントディレクトリの `.env` を優先的に読み込むように `config.go` を修正しました。

2.  **APIプロキシ設定の欠落 (Frontend)**
    - `vite.config.js` にプロキシ設定を追加し、開発環境 (`npm run dev`) で `/api` リクエストがバックエンド (`localhost:8080`) に転送されるようになりました。

3.  **uploadsディレクトリの欠落 (Backend)**
    - 静的ファイル配信用の `uploads` ディレクトリを作成しました。
    - **Note**: `backapp/.gitignore` に登録済みです。
