# SportEase 開発者仕様書

## 1. システム概要

SportEaseは、学校のスポーツイベントを管理するためのWebアプリケーションです。以下の2つの主要コンポーネントで構成されています。
- **Frontapp**: SvelteKitベースのアプリケーション。画面の提供に加え、`/api` をバックエンドへ中継する信頼境界でもある。
- **Backapp**: Go (Gin) ベースのREST APIサーバー。MySQLに加え、Redisでセッション・CSRFトークン・ユーザー単位のレート制限を管理する。

## 2. アーキテクチャ

### Frontapp (SvelteKit)
- 場所: `/frontapp`
- 言語: JavaScript/Svelte
- 開発サーバーポート: 5173 (デフォルト)
- 主な依存関係: `svelte`, `vite`, `tailwindcss`

### Backapp (Go)
- 場所: `/backapp`
- 言語: Go 1.26（`backapp/go.mod` を正とする）
- サーバーポート: 8080
- データベース: MySQL 8（`database/sql` + MySQLドライバー）
- 認証: Google OAuth 2.0 / OpenID Connect。IDトークンを検証してログインする。
- セッション: Redisに24時間保存。ブラウザにはHttpOnlyの`session_token`とCSRFトークンCookieを発行する。

## 3. セットアップ手順 (ローカル開発)

### 前提条件
- Go 1.26系（`backapp/go.mod` の `toolchain` 指定に従う）
- Node.js 18以上
- Docker (オプション、データベース用)

### 手順 1: バックエンドのセットアップ
1.  `backapp` ディレクトリに移動します。
2.  プロジェクトルートの `.env` ファイルを作成します。バックエンドは起動ディレクトリに関係なく、プロジェクトルート直下の `.env` を読み込みます。
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

### 認証・更新系APIの扱い

- OAuth開始・コールバックは、IPベースで1分あたり10回に制限されます。`state`と`nonce`は一度使うと失効します。
- `/api` の安全でないHTTPメソッド（`POST`、`PUT`、`DELETE`など）は、セッションCookieを送る場合にCSRFトークンを要求します。通常のブラウザアクセスでは、SvelteKitプロキシが同一オリジンを確認してトークンを転送します。
- サーバー側からバックエンドへ更新系APIを直接呼ぶ場合は、`session_token`と`csrf_token`をCookieに含め、`X-CSRF-Token`へ同じCSRFトークンを設定します。フロントの`createBackendSessionHeaders`を利用するとこの形式になります。
- プロキシを追加・変更する場合は、バックエンドの`TRUSTED_PROXY_CIDRS`をそのプロキシのCIDRへ限定します。クライアントから渡された`X-Forwarded-*`をそのままバックエンドへ中継してはいけません。

### Push購読の扱い

- `POST /api/notifications/subscription`はユーザーごとに1時間10回までです。
- 購読情報は最大4 KiB、ユーザーあたり最大5件です。HTTPS/443かつ`WEBPUSH_ALLOWED_HOSTS`（未設定時は主要ブラウザの既定Pushサービス）に一致するエンドポイントだけを保存できます。
- Push送信は最大32件並行、全体30秒のバッチ期限で実行します。410/404または不正な購読情報は削除対象です。

## 5. 既知の問題とバグ
現在のコードベースで確認されている問題の詳細なリストについては、[bugs.md](./bugs.md) を参照してください。
