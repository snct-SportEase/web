# SportEase

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/snct-SportEase/web)

SportEaseは、学校行事のスポーツ大会運営をオンラインで一元管理するためのWebアプリケーションです。Google認証を利用した安全なログイン、役割に応じたダッシュボード、トーナメント進行管理、通知配信、出席確認など、運営と参加者双方の体験を支援します。

## 主な機能

### 認証・共通
- Google OAuthによるサインインと学校メールドメイン制御（`sendai-nct.jp` / `sendai-nct.ac.jp`）
- 役割に応じたダッシュボード表示（root / admin / student）
- クラス別の参加状況・スコア・進行状況の可視化
- Web Push通知の購読管理とPWA対応

### Root（システム管理者）
- 大会（イベント）の作成・更新・ステータス管理（準備中・予定・開催中・アーカイブ）
- トーナメント一括生成、プレビュー、ノーンゲーム設定管理
- 競技の登録、チーム一覧の参照
- クラス在籍人数の更新（CSVインポート対応）
- 通知の作成・配信対象ロールの管理
- 通知申請の審査・メッセージや決裁結果の記録
- `admin` / `root` 基本権限の付与・剥奪
- ユーザー表示名の管理
- MIC対象クラスの集計、ポイント調整

### Admin（大会運営担当）
- クラス・チーム編成とメンバー割当
- 審判ロールなど任意ロールの付与・削除
- 競技詳細情報や競技要項PDFのアップロード
- 出席登録とクラス別出席状況の参照
- 試合開始時刻・進行ステータスの更新、開催中大会の試合結果入力
- 開催中大会のノーンゲーム試合結果登録、MIC投票
- MyIDバーコード読み取りによる参加本登録・ラウンドチェックイン

### 大会ステータスの運用

| ステータス | 用途 |
| --- | --- |
| 準備中 (`preparing`) | 現在操作する大会として選択され、競技設定・チーム編成・トーナメント生成などの準備を行う状態です。試合結果は入力できません。 |
| 予定 (`upcoming`) | 開催前の大会情報を保管する状態です。操作対象にはなりません。 |
| 開催中 (`active`) | 大会当日の運用状態です。試合結果・昼競技結果を入力できます。 |
| アーカイブ (`archived`) | 終了した大会を参照用に保管する状態です。 |

### Student（参加者）
- マイページでの自身の競技参加状況とスコア履歴の閲覧
- クラス全体の出席・勝ち進み状況の確認
- 通知一覧と通知申請フォーム、root宛てのメッセージ送信
- 競技ルールPDF・資料の閲覧
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
| `GOOGLE_REDIRECT_URL` | Google OAuthコールバックURL（開発環境の例: `http://localhost:3300/api/auth/google/callback`） |
| `FRONTEND_URL` | フロントエンドのベースURL（開発環境の例: `http://localhost:3300`） |
| `INIT_ROOT_USER` | 初回ログイン時にroot権限を付与する初期rootユーザーのメールアドレス |
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
ブラウザから `http://localhost:5000` にアクセスし、Googleアカウントでログインします。`sendai-nct.jp` または `sendai-nct.ac.jp` ドメイン以外のメールアドレスは拒否されます。許可されたドメインで初めてログインしたユーザーは自動的にstudentとして登録され、`INIT_ROOT_USER` と一致するユーザーにはroot権限が付与されます。

### 4. Docker Composeでの起動
初回起動時、またはDBスキーマ変更を取り込む場合は、先にマイグレーションを実行します。

```bash
cd /home/saku0512/Desktop/develop/SportEase/webapp
docker compose --profile migration run --rm migrate
docker compose up --build
```
開発用Traefikは3300番ポートだけを公開します。`http://localhost:3300` でフロントエンドと `/api` プレフィックスのバックエンドを利用でき、HTTPSリダイレクトは行いません。80/443番とLet's Encryptを使う本番設定は `docker-compose.production.yml` に分離されています。

#### PWA・Push通知の状態表示

通知を利用できる student・admin・root ユーザーには、ダッシュボードのヘッダーとサイドバーに現在の設定状態が表示されます。

- `PWA未設定`: 現在PWAとして起動していない場合に表示されます。クリックすると導入画面が開き、対応するChromium系ブラウザではブラウザ標準のインストール確認を呼び出します。標準APIを利用できないブラウザではOS別の手動インストール手順を表示します。
- `通知未設定` / `通知拒否中`: 現在の端末でPush通知を受信できない場合に表示されます。クリックするとダッシュボードのPush通知設定へ移動します。

ブラウザから端末上のPWAインストール状況を常に取得できる標準APIはないため、`PWA未設定` は現在の表示モードが `standalone` かどうかを基準に判定します。

### 5. 開発用デモデータ

Docker Compose の開発環境へ、競技・チーム・トーナメントなどの検証用データを投入できます。

```bash
docker compose -f docker-compose.yml up -d --build
docker compose -f docker-compose.yml --profile demo run --rm demo-data
```

`demo-data` は既存DBのスキーマを検出して必要に応じてマイグレーションのベースラインを登録し、未適用のDBマイグレーションを実行してから、2037年度春季の「デモ体育大会」をアクティブな大会として登録します。主なデータは次のとおりです。

- 通常運用と同じ16クラス、各クラス8名のデモ生徒、root・adminユーザー
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
