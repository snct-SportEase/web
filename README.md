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
- `admin` / `root` 基本権限の付与・剥奪
- ユーザー表示名とクラス所属ロール（`クラス名_rep`）の管理
- MIC対象クラスの集計、ポイント調整

### Admin（大会運営担当）
- クラス・チーム編成とメンバー割当
- 審判ロールなど任意ロールの付与・削除
- 競技詳細情報や資料（画像・PDF）のアップロード
- 出席登録とクラス別出席状況の参照
- 試合開始時刻・進行ステータス・結果の更新
- ノーンゲーム試合結果登録、MIC投票
- MyIDバーコード読み取りによる参加本登録・ラウンドチェックイン

### Student（参加者）
- マイページでの自身の競技参加状況とスコア履歴の閲覧
- クラス全体の出席・勝ち進み状況の確認
- 通知一覧と通知申請フォーム、root宛てのメッセージ送信
- 競技ルール・資料の閲覧
- 大会当日のMyIDバーコード提示による参加確認

## 技術スタック
- フロントエンド: SvelteKit, Vite, Tailwind CSS, Playwright, Vitest
- バックエンド: Go 1.24 + Gin, MySQL, WebSocket, WebPush
- インフラ: Docker Compose, Traefik, Let's Encrypt

## ディレクトリ構成
- `frontapp/` フロントエンド（SvelteKit）ソースとテスト
- `backapp/` バックエンド（Go）ソース、DBマイグレーション、テスト
- `docker-compose.yml` 本番／ステージング想定のコンテナ定義

メンテナンス時に「どこに何のコードがあるか」を探す場合は、[コード配置・保守ガイド](./docs/maintenance_code_map.md) を参照してください。

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
MySQLを別途用意し、プロジェクトルートの `.env` に記載した接続情報で参照できるようにしてください。初回起動時に初期rootユーザー・イベント・クラス情報が自動作成されます。

#### フロントエンド（SvelteKit）
```bash
cd /home/saku0512/Desktop/develop/SportEase/webapp/frontapp
npm install
npm run dev -- --host --port 5000
```
ブラウザから `http://localhost:5000` にアクセスし、Googleアカウントでログインします。ホワイトリストに登録されていないメールアドレスは拒否されます。

### 4. Docker Composeでの起動
初回起動時、またはDBスキーマ変更を取り込む場合は、先にマイグレーションを実行します。

```bash
cd /home/saku0512/Desktop/develop/SportEase/webapp
docker compose --profile migration run --rm migrate
docker compose up --build
```
Traefikが80/443/3300番ポートを公開し、`localhost` でフロントエンド（3300番エントリポイント）、`/api` プレフィックスでバックエンドが利用可能になります。

### 5. 開発用デモデータ

Docker Compose の開発環境へ、競技・チーム・トーナメントなどの検証用データを投入できます。

```bash
docker compose -f docker-compose.yml up -d --build
docker compose -f docker-compose.yml --profile demo run --rm demo-data
```

`demo-data` は既存DBのスキーマを検出して必要に応じてマイグレーションのベースラインを登録し、未適用のDBマイグレーションを実行してから、2037年度春季の「デモ体育大会」をアクティブな大会として登録します。主なデータは次のとおりです。

- 通常運用と同じ16クラス、各クラス8名のデモ生徒、root・admin・全生徒のクラスロール
- バスケットボール・バレーボール・サッカーの全クラス分のチーム
- 終了済み・開始予定・未確定の試合を含むトーナメント
- 出席、試合ラウンド受付、MIC投票、雨天時設定
- 昼競技のグループ、試合、結果、順位詳細、得点
- クラス得点、対象ロール・種別の異なる通知、各状態の通知申請と会話履歴
- 大会要項とroot・admin・student向けPDFガイド

このコマンドは再実行できます。再実行するとデモ大会の対戦表と得点は初期状態に戻ります。開発・検証専用であり、`docker-compose.production.yml` からは実行できません。デモユーザーのメールアドレスはデータ関連の画面確認用で、認証自体を迂回しないためログインには従来どおり Google OAuth が必要です。

## DBマイグレーション

DBスキーマは `golang-migrate` 形式のSQLで管理します。マイグレーションファイルは `backapp/db/migrations/` に配置し、`000001_xxx.up.sql` / `000001_xxx.down.sql` のペアで追加します。

通常の `docker compose up -d` ではマイグレーションは自動実行されません。明示的に実行する場合は `migration` profile を指定します。

```bash
docker compose -f docker-compose.production.yml --profile migration run --rm migrate
```

適用済みバージョンはDB内の `schema_migrations` テーブルで確認できます。

```bash
docker compose -f docker-compose.production.yml exec sportease-db \
  mysql -u root -p"$DB_ROOT_PASSWORD" "$DB_DATABASE" \
  -e "SELECT * FROM schema_migrations;"
```

### 本番DBへの反映手順

まず最新コードを取得します。

```bash
git pull origin main
```

本番DBをバックアップします。

```bash
mkdir -p ~/db-backups

docker compose -f docker-compose.production.yml exec sportease-db \
  mysqldump -u root -p"$DB_ROOT_PASSWORD" "$DB_DATABASE" \
  > ~/db-backups/sportease_before_migrate_$(date +%Y%m%d_%H%M%S).sql
```

既存本番DBに初めて `golang-migrate` を導入する場合、すでに手動適用済みのスキーマを再実行しないように、初回だけベースラインを登録します。`000001_initial_schema` と `000002_add_guide_documents` が適用済みで、`000003_add_round_check_ins` が未適用のDBでは version `2` を登録します。

```bash
docker compose -f docker-compose.production.yml exec sportease-db \
  mysql -u root -p"$DB_ROOT_PASSWORD" "$DB_DATABASE" \
  -e "CREATE TABLE IF NOT EXISTS schema_migrations (version bigint not null primary key, dirty boolean not null); REPLACE INTO schema_migrations (version, dirty) VALUES (2, false);"
```

未適用のマイグレーションだけを適用します。

```bash
docker compose -f docker-compose.production.yml --profile migration run --rm migrate
```

適用状況を確認します。

```bash
docker compose -f docker-compose.production.yml exec sportease-db \
  mysql -u root -p"$DB_ROOT_PASSWORD" "$DB_DATABASE" \
  -e "SELECT * FROM schema_migrations;"
```

最後にアプリを更新します。

```bash
docker compose -f docker-compose.production.yml up -d --build
```

`down` はデータ削除を伴う可能性があるため、本番環境では緊急時以外実行しないでください。

## テスト実行
- フロントエンド: `npm run test`（単体テスト）、`npm run test:e2e`（Playwright）
- バックエンド: `go test ./...`

## 補足
- トーナメント進行状況はWebSocket (`/api/ws/tournaments/:tournament_id`) で配信されます。
- 画像やPDF資料はバックエンドの`/uploads`ディレクトリに保存され、Traefik経由で配信されます。
- Push通知を有効にする場合は、Service Worker登録とHTTPS環境が必要です。
