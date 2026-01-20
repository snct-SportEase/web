# SportEase 開発者仕様書

## 1. システム概要

SportEaseは、学校のスポーツイベントを管理するためのWebアプリケーションです。以下の2つの主要コンポーネントで構成されています。
- **Frontapp**: SvelteKitベースのシングルページアプリケーション (SPA)。
- **Backapp**: Go (Gin) ベースのREST APIサーバー。

## 2. アーキテクチャ

### Frontapp (SvelteKit)
- 場所: `/frontapp`
- 言語: JavaScript/Svelte
- 開発サーバーポート: 5173 (デフォルト)
- 主な依存関係: `svelte`, `vite`, `tailwindcss`

### Backapp (Go)
- 場所: `/backapp`
- 言語: Go
- サーバーポート: 8080
- データベース: `database/sql` を使用 (使用状況からMySQL/PostgreSQLベースと推測されます)。
- 認証: Google OAuth2

## 3. セットアップ手順 (ローカル開発)

### 前提条件
- Go 1.21以上
- Node.js 18以上
- Docker (オプション、データベース用)

### 手順 1: バックエンドのセットアップ
1.  `backapp` ディレクトリに移動します。
2.  `.env` ファイルを作成します。 **注意**: アプリケーションは現在、実行パスからの相対パスで親ディレクトリなどにある `.env` を探そうとするバグがありましたが、カレントディレクトリも探すように修正されました。`backapp` 直下、または `web` 直下に `.env` を配置することをお勧めします。
3.  `backapp` ルートに `uploads` ディレクトリが存在することを確認します（修正済み）。
    ```bash
    mkdir -p backapp/uploads
    ```
4.  サーバーを起動します。
    ```bash
    cd backapp
    go run cmd/server/main.go
    ```

### 手順 2: フロントエンドのセットアップ
1.  `frontapp` ディレクトリに移動します。
2.  依存関係をインストールします。
    ```bash
    npm install
    ```
3.  `vite.config.js` で `/api` リクエストをバックエンドにプロキシするように設定します（修正済み）。
4.  開発サーバーを起動します。
    ```bash
    npm run dev
    ```

## 4. API構造
バックエンドは `/api` 以下にエンドポイントを公開しています。
- `/api/auth/*`: 認証 (Googleログイン)
- `/api/events/*`: イベント管理
- `/api/users/*`: ユーザー管理
- `/api/notifications/*`: プッシュ通知

## 5. 既知の問題とバグ
現在のコードベースで確認されている問題の詳細なリストについては、[bugs.md](./bugs.md) を参照してください。
