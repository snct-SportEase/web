# SportEase 開発者仕様書

## 1. システム概要

SportEaseは、学校のスポーツイベントを管理するためのWebアプリケーションです。以下の2つの主要コンポーネントで構成されています。
- **Frontapp**: SvelteKitベースのアプリケーション。画面の提供に加え、`/api` をバックエンドへ中継する信頼境界でもある。
- **Backapp**: Go (Gin) ベースのREST APIサーバー。MySQLに加え、Redisでセッション・CSRFトークン・ユーザー単位のレート制限を管理する。

## 2. アーキテクチャ

### Frontapp (SvelteKit)
- 場所: `/frontapp`
- 言語: JavaScript/Svelte
- 開発サーバーポート: 5000（`frontapp/package.json` の `npm run dev`）
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
3.  初回だけ、アップロードファイル保存用ディレクトリを作成します。
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
3.  開発サーバーを起動します。`vite.config.js` の設定により `/api` はバックエンドへプロキシされます。
    ```bash
    npm run dev
    ```

## 4. API構造
バックエンドは `/api` 以下にエンドポイントを公開しています。
- `/api/auth/*`: 認証 (Googleログイン)
- `/api/events/*`: 現在の大会、競技、チーム、通常トーナメント、盤上競技、昼競技の参照
- `/api/admin/*`: 出席、参加登録、試合結果、昼競技結果、競技詳細、統計、任意ロール管理
- `/api/root/*`: 大会・競技マスター、通常/盤上トーナメント生成、昼競技テンプレート、通知、権限、バックアップ
- `/api/users/*`: 自分のプロフィールとユーザー情報
- `/api/notifications/*`: アプリ内通知、Push購読、通知申請
- `/api/ws/tournaments/:id`: 通常トーナメントのリアルタイム更新

正確なルート、適用middleware、権限は `backapp/internal/router/router.go` を正とします。主な追加データフローは次の通りです。

- 通常競技の一括生成は将棋・オセロを除外し、盤上競技は `board_game_*` テーブルと専用APIで管理します。
- 昼競技は1大会に複数セッションを持ち、テンプレートキーでリレー、綱引き、競技タイピング、借り物競争などを識別します。
- 競技タイピングは `typing-results-v1` JSONを取り込み、インポート履歴・結果・大会得点を同一処理で保存します。
- 競技参加ロールは `event_id` を持ち、大会間で通知対象や割り当てが混在しないようにします。
- 通知はロール宛てに加えて個人ユーザー宛ても指定できます。

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
