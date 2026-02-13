# SportEase

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/snct-SportEase/web)

SportEaseは、学校行事のスポーツ大会運営をオンラインで一元管理するためのWebアプリケーションです。Google認証を利用した安全なログイン、役割に応じたダッシュボード、トーナメント進行管理、通知配信、出席確認など、運営と参加者双方の体験を支援します。

## 主な機能

### 認証・共通
- Google OAuthによるサインインとメールホワイトリスト制御
- 役割に応じたダッシュボード表示（root / admin / student）
- クラス別の参加状況・スコア・進行状況の可視化
- Web Push通知の購読管理とPWA対応

### Root（システム管理者）
- 大会（イベント）の作成・更新・アクティブ切り替え
- トーナメント一括生成、プレビュー、ノーンゲーム設定管理
- 競技の登録、チーム一覧の参照
- クラス在籍人数の更新（CSVインポート対応）
- 通知の作成・配信対象ロールの管理
- 通知申請の審査・メッセージや決裁結果の記録
- ログイン許可メールアドレスの管理（単体／CSV登録）
- MVP対象クラスの集計、ポイント調整

### Admin（大会運営担当）
- クラス・チーム編成とメンバー割当
- ロール（権限）の付与・削除、表示名の一括変更
- 競技詳細情報や資料（画像・PDF）のアップロード
- 出席登録とクラス別出席状況の参照
- 試合開始時刻・進行ステータス・結果の更新
- ノーンゲーム試合結果登録、MVP投票
- QRコードの生成・検証ツール

### Student（参加者）
- マイページでの自身の競技参加状況とスコア履歴の閲覧
- クラス全体の出席・勝ち進み状況の確認
- 通知一覧と通知申請フォーム、root宛てのメッセージ送信
- 競技ルール・資料の閲覧
- 自身の参加証QRコード発行

## 技術スタック
- フロントエンド: SvelteKit, Vite, Tailwind CSS, Playwright, Vitest
- バックエンド: Go 1.24 + Gin, MySQL, WebSocket, WebPush
- インフラ: Docker Compose, Traefik, Let's Encrypt

## ディレクトリ構成
- `frontapp/` フロントエンド（SvelteKit）ソースとテスト
- `backapp/` バックエンド（Go）ソース、DBマイグレーション、テスト
- `docker-compose.yml` 本番／ステージング想定のコンテナ定義

## セットアップ

### 1. 前提ツール
- Node.js 18 以上（推奨 LTS）
- npm
- Go 1.24 系
- MySQL 8 系（Docker使用時は不要）
- OpenSSL（Web Push鍵の生成に使用）

### 2. 環境変数

#### プロジェクトルート `.env`（バックエンド・Docker共通）
| 変数 | 内容 |
| --- | --- |
| `DB_HOST` | MySQLホスト名（例: `127.0.0.1`） |
| `DB_PORT` | MySQLポート（例: `3306`） |
| `DB_USER` | アプリ用DBユーザー |
| `DB_PASSWORD` | アプリ用DBパスワード |
| `DB_DATABASE` | 使用するデータベース名 |
| `DB_ROOT_PASSWORD` | docker-composeでMySQLを起動する場合のrootパスワード |
| `GOOGLE_CLIENT_ID` | Google OAuth クライアントID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth クライアントシークレット |
| `GOOGLE_REDIRECT_URL` | Google OAuthコールバックURL（例: `https://localhost/api/auth/google/callback`） |
| `FRONTEND_URL` | フロントエンドのベースURL（例: `https://localhost`） |
| `INIT_ROOT_USER` | 初期rootユーザーとしてホワイトリスト登録するメールアドレス |
| `INIT_EVENT_NAME` | 初期イベント名 |
| `INIT_EVENT_YEAR` | 初期イベントの年度（例: `2025`） |
| `INIT_EVENT_SEASON` | `spring` または `autumn` |
| `INIT_EVENT_START_DATE` | イベント開始日（`YYYY-MM-DD`） |
| `INIT_EVENT_END_DATE` | イベント終了日（`YYYY-MM-DD`） |
| `WEBPUSH_PUBLIC_KEY` | Web PushのVAPID公開鍵（Base64, URL Safe） |
| `WEBPUSH_PRIVATE_KEY` | Web PushのVAPID秘密鍵 |
| `LETSENCRYPT_EMAIL` | Traefik用のLet's Encrypt通知メールアドレス |

> `WEBPUSH_*` は `openssl` 等でVAPID鍵を生成して設定してください。開発中にPush通知を使用しない場合は未設定でも動作しますが、対応機能は無効化されます。

#### フロントエンド `frontapp/.env`
| 変数 | 内容 |
| --- | --- |
| `BACKEND_URL` | バックエンドAPIのベースURL（例: `http://localhost:8080`） |
| `PUBLIC_WEBPUSH_PUBLIC_KEY` | Web PushのVAPID公開鍵。通知を利用しない場合は空でも可 |

### 3. ローカル開発手順

#### バックエンド（Go）
```bash
cd /home/saku0512/Desktop/develop/SportEase/webapp/backapp
go mod download
go run cmd/server/main.go
```
MySQLを別途用意し、`.env` に記載した接続情報で参照できるようにしてください。初回起動時に初期rootユーザー・イベント・クラス情報が自動作成されます。

#### フロントエンド（SvelteKit）
```bash
cd /home/saku0512/Desktop/develop/SportEase/webapp/frontapp
npm install
npm run dev -- --host --port 5000
```
ブラウザから `http://localhost:5000` にアクセスし、Googleアカウントでログインします。ホワイトリストに登録されていないメールアドレスは拒否されます。

### 4. Docker Composeでの起動
```bash
cd /home/saku0512/Desktop/develop/SportEase/webapp
docker compose up --build
```
Traefikが80/443/3300番ポートを公開し、`localhost` でフロントエンド（3300番エントリポイント）、`/api` プレフィックスでバックエンドが利用可能になります。

## テスト実行
- フロントエンド: `npm run test`（単体テスト）、`npm run test:e2e`（Playwright）
- バックエンド: `go test ./...`

## 補足
- トーナメント進行状況はWebSocket (`/api/ws/tournaments/:tournament_id`) で配信されます。
- 画像やPDF資料はバックエンドの`/uploads`ディレクトリに保存され、Traefik経由で配信されます。
- Push通知を有効にする場合は、Service Worker登録とHTTPS環境が必要です。

