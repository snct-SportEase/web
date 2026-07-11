# SportEase コード配置・保守ガイド

このドキュメントは、SportEase をメンテナンスするときに「どこに何のコードがあるか」を素早く探すための地図です。詳細な仕様は各コードと既存ドキュメントを正とし、ここでは入口になるファイルと変更時の見方をまとめます。

## 全体像

| 場所 | 役割 | 主に見るタイミング |
| --- | --- | --- |
| `frontapp/` | SvelteKit のフロントエンド | 画面、フォーム、表示、PWA、フロントのテストを直すとき |
| `backapp/` | Go + Gin のバックエンド API | API、認証、DB操作、WebSocket、ファイルアップロードを直すとき |
| `backapp/db/migrations/` | DBスキーマのマイグレーション | テーブル・カラム・制約を追加変更するとき |
| `docs/` | 開発・運用ドキュメント | 手順や仕様、テスト方針、既知バグを確認するとき |
| `latex/` | 利用説明などのPDF/LaTeX資料 | 配布資料や手順書PDFを更新するとき |
| `docker-compose.yml` | ローカル/検証向け Docker Compose | Dockerで起動構成を変更するとき |
| `docker-compose.production.yml` | 本番向け Docker Compose | 本番のTraefik、環境変数、コンテナ設定を変更するとき |

## まず見るファイル

| ファイル | 内容 |
| --- | --- |
| `README.md` | セットアップ、環境変数、Docker、DBマイグレーション、テストの入口 |
| `docs/developer_spec.md` | 開発者向けの概要仕様 |
| `docs/go_test_guidelines.md` | Goテストの書き方・実行方針 |
| `docs/frontend_test_guidelines.md` | フロントエンドテストの書き方・実行方針 |
| `docs/user_permission_guide.md` | ユーザー権限まわりの説明 |
| `docs/bugs.md` | 既知の問題 |
| `NOTIFICATION_TROUBLESHOOTING.md` | 通知まわりのトラブルシュート |
| `SECURITY.md` | セキュリティ方針 |

## フロントエンド

### 基本構成

| 場所 | 内容 |
| --- | --- |
| `frontapp/src/routes/` | 画面とSvelteKitのルーティング。URLとほぼ同じ構造 |
| `frontapp/src/routes/+page.svelte` | ログイン前トップページ |
| `frontapp/src/routes/+layout.svelte` | アプリ全体の共通レイアウト |
| `frontapp/src/routes/+layout.server.js` | ルートレイアウトのサーバー側ロード処理 |
| `frontapp/src/routes/dashboard/+layout.svelte` | ログイン後ダッシュボードの共通レイアウト、ヘッダー、フッター、PWA通知導線 |
| `frontapp/src/routes/dashboard/+layout.server.js` | ダッシュボード共通のユーザー情報受け渡し |
| `frontapp/src/routes/dashboard/+page.svelte` | ダッシュボードトップ |
| `frontapp/src/lib/components/` | 複数画面で使うSvelteコンポーネント |
| `frontapp/src/lib/stores/` | Svelteストア。サイドバー、イベント、通知バッジ、Push購読状態など |
| `frontapp/src/lib/utils/` | PWA、Push通知、HTMLサニタイズなどの共通処理 |
| `frontapp/src/app.css` | グローバルCSS |
| `frontapp/src/app.html` | HTMLテンプレート |
| `frontapp/src/service-worker.js` | PWAキャッシュとPush通知受信 |
| `frontapp/static/` | manifestやアイコンなど静的ファイル |

### 役割別画面

| 画面の種類 | 場所 | 主な内容 |
| --- | --- | --- |
| root向け | `frontapp/src/routes/dashboard/root/` | 大会、競技、通知、通知申請、雨天モード、トーナメント、ユーザー昇格、MIC(行事委員会賞)、資料アップロード |
| admin向け | `frontapp/src/routes/dashboard/admin/` | 出席、バーコード読み取り、クラス管理、参加者確定、試合結果入力、競技詳細、MIC(行事委員会賞)投票 |
| student向け | `frontapp/src/routes/dashboard/student/` | マイページ、クラス情報、通知、通知申請、昼競技、スコア、競技情報、時間割、トーナメント |
| 共通/補助 | `frontapp/src/routes/dashboard/archive/` | 過去イベントの閲覧 |
| 共通/補助 | `frontapp/src/routes/dashboard/guide/` | 競技ガイド資料の閲覧 |
| 共通/補助 | `frontapp/src/routes/dashboard/privacy-policy/` | プライバシーポリシー |

### フロントのAPIプロキシとサーバー処理

| 場所 | 内容 |
| --- | --- |
| `frontapp/vite.config.js` | `/api` をバックエンドへプロキシ。テスト設定もここにある |
| `frontapp/src/routes/api/notifications/+server.js` | フロント側の通知API補助 |
| `frontapp/src/routes/api/notifications/subscription/+server.js` | Push購読に関するフロント側API補助 |
| 各 `+page.server.js` | 対象ページのサーバー側データ取得や権限確認 |

## バックエンド

### 基本構成

| 場所 | 内容 |
| --- | --- |
| `backapp/cmd/server/main.go` | アプリ起動、設定読み込み、DB接続、Redisセッション初期化、初期イベント/クラス/スコア作成、HTTPサーバー起動 |
| `backapp/internal/config/config.go` | 環境変数と `.env` 読み込み |
| `backapp/internal/router/router.go` | APIルート定義の中心。どのURLがどのhandlerを呼ぶかはここを見る |
| `backapp/internal/handler/` | HTTPリクエスト/レスポンス処理、入力検証、権限前提のユースケース処理 |
| `backapp/internal/repository/` | DBアクセス。SQLやトランザクション、永続化ロジック |
| `backapp/internal/repository/db.go` | DB接続設定、接続プール、`GlobalCache`、`singleflight`、`GlobalAnts` などの共通最適化リソース |
| `backapp/internal/models/` | APIやDBで扱う構造体 |
| `backapp/internal/middleware/` | 認証、権限、CORS、レート制限 |
| `backapp/internal/websocket/` | WebSocket Hub、接続クライアント、進行状況配信 |
| `backapp/db/migrations/` | `golang-migrate` 用SQL |
| `backapp/db/conf.d/custom.cnf` | MySQL設定 |

### APIルートの読み方

APIの入口は `backapp/internal/router/router.go` です。保守時は次の順で追うと迷いにくいです。

1. 変更したいURLや画面が呼ぶ `/api/...` を探す。
2. `router.go` で該当ルートとhandler名を確認する。
3. `backapp/internal/handler/*_handler.go` でリクエスト処理を見る。
4. DBを読む/書く処理は `backapp/internal/repository/*_repository.go` に降りる。
5. 返却データや入力構造体は `backapp/internal/models/` を確認する。
6. DBスキーマ変更が必要なら `backapp/db/migrations/` にup/downを追加する。
7. 対応するテストを `backapp/tests/` または `frontapp/tests/` に追加・更新する。

### 認証・権限

| 場所 | 内容 |
| --- | --- |
| `backapp/internal/handler/auth_handler.go` | Google OAuth、ログイン、ログアウト、ユーザー取得、プロフィール更新、ユーザー管理 |
| `backapp/internal/middleware/auth.go` | `session_token` Cookie、Redisセッション、`AuthMiddleware`、`RoleRequired`、クラス所属権限 |
| `backapp/internal/repository/user_repository.go` | ユーザー、ロール、ホワイトリスト関連のDB処理 |
| `frontapp/src/routes/dashboard/+layout.server.js` | ログイン後画面へユーザー情報を渡す入口 |
| `frontapp/src/routes/dashboard/root/user-promotion/` | rootによるユーザー昇格・降格画面 |
| `frontapp/src/lib/components/ProfileSetupModal.svelte` | 初回プロフィール設定 |
| `frontapp/src/lib/components/EditDisplayNameModal.svelte` | 表示名変更 |
| `docs/user_permission_guide.md` | 権限仕様の補足 |

## 機能別の主な関連ファイル

| 機能 | フロント | バックエンドhandler | repository/model | テストの場所 |
| --- | --- | --- | --- | --- |
| イベント管理（重複登録のクラス人数上限を含む） | `frontapp/src/routes/dashboard/root/event-management/` | `event_handler.go` | `event_repository.go`, `event.go`, `db/migrations/000005_add_duplicate_registration_threshold.*.sql` | `backapp/tests/handler/event_handler_test.go`, `backapp/tests/repository/event_repository_test.go`, `frontapp/tests/e2e/root-event-management.spec.js` |
| 競技管理（競技ルールはPDFのみ） | `frontapp/src/routes/dashboard/root/sport-management/`, `frontapp/src/routes/dashboard/admin/sport-details-registration/`, `frontapp/src/routes/dashboard/student/sport-info/` | `sport_handler.go`, `pdf_handler.go` | `sport_repository.go`, `sport.go` | `backapp/tests/handler/sport_handler_test.go`, `backapp/tests/repository/sport_repository_test.go`, `frontapp/tests/e2e/root-sport-management.spec.js` |
| トーナメント管理 | `frontapp/src/routes/dashboard/root/tournament-management/`, `frontapp/src/routes/dashboard/student/tournament/` | `tournament_handler.go`, `tournament_export_handler.go`, `all_tournament_handler.go` | `tournament_repository.go`, `team_repository.go`, `tournament.go`, `team.go` | `backapp/tests/handler/all_tournament_handler_test.go`, `backapp/tests/repository/tournament_repository_test.go`, `frontapp/tests/e2e/root-tournament-management.spec.js` |
| 試合結果入力 | `frontapp/src/routes/dashboard/admin/insert-matche-result/`, `frontapp/src/lib/components/InsertMatchResultModal.svelte`, `frontapp/src/lib/components/ConfirmMatchResultModal.svelte` | `tournament_handler.go` | `tournament_repository.go`, `class_score_repository.go` | `backapp/tests/handler/tournament_export_handler_test.go` など |
| 昼競技 | `frontapp/src/routes/dashboard/root/noon-game/`, `frontapp/src/routes/dashboard/admin/noon-game-results/`, `frontapp/src/routes/dashboard/student/noon-game/` | `noon_game_handler.go` | `noon_game_repository.go`, `noon_game.go` | `backapp/tests/handler/noon_game_*.go`, `frontapp/tests/e2e/root-noon-game.spec.js` |
| 雨天モード | `frontapp/src/routes/dashboard/root/rainy-mode/` | `rainy_mode_handler.go`, `event_handler.go` | `rainy_mode_repository.go`, `rainy_mode_setting.go` | `backapp/tests/handler/rainy_mode_handler_test.go`, `backapp/tests/repository/rainy_mode_repository_test.go`, `frontapp/tests/e2e/root-rainy-mode.spec.js` |
| クラス・チーム管理（重複登録判定を含む） | `frontapp/src/routes/dashboard/admin/class-management/`, `frontapp/src/routes/dashboard/student/class-info/` | `class_handler.go`, `class_team_handler.go` | `class_repository.go`, `team_repository.go`, `class.go`, `team.go`, `event.go` | `backapp/tests/handler/class_handler_test.go`, `backapp/tests/handler/class_team_handler_test.go`, `backapp/tests/repository/class_repository_test.go`, `backapp/tests/repository/team_repository_test.go` |
| クラス在籍人数 | `frontapp/src/routes/dashboard/root/class-student-count/` | `class_handler.go` | `class_repository.go` | `backapp/tests/handler/class_handler_export_test.go`, `frontapp/tests/e2e/root-class-student-count.spec.js` |
| 出席管理 | `frontapp/src/routes/dashboard/admin/attendance-management/` | `attendance_handler.go` | `class_repository.go`, `event_repository.go` | `backapp/tests/handler/attendance_handler_test.go`, `frontapp/tests/e2e/admin-barcode-check-in.spec.js` |
| バーコード/MyID | `frontapp/src/routes/dashboard/admin/barcode-reader/`, `frontapp/src/routes/dashboard/admin/confirmed-participants/` | `barcode_handler.go` | `team_repository.go`, `sport_repository.go`, `user_repository.go` | `backapp/tests/handler/barcode_handler_test.go`, `frontapp/tests/e2e/admin-barcode-check-in.spec.js` |
| 通知配信 | `frontapp/src/routes/dashboard/root/notification/`, `frontapp/src/routes/dashboard/student/notification/`, `frontapp/src/lib/components/NotificationSettings.svelte` | `notification_handler.go`, `event_handler.go`, `push_subscription.go` | `notification_repository.go`, `notification.go` | `backapp/tests/handler/notification_handler_test.go`, `frontapp/tests/e2e/root-notification.spec.js` |
| 通知申請 | `frontapp/src/routes/dashboard/root/notification-requests/`, `frontapp/src/routes/dashboard/student/notification-request/` | `notification_request_handler.go` | `notification_request_repository.go`, `notification_request.go` | `backapp/tests/handler/notification_handler_test.go`, `frontapp/tests/e2e/root-notification-requests.spec.js` |
| PWA/Push購読 | `frontapp/src/service-worker.js`, `frontapp/src/lib/utils/push.js`, `frontapp/src/lib/utils/pwa.js`, `frontapp/src/lib/stores/pushSubscriptionStore.js`, `frontapp/src/lib/stores/pwaInstallStore.js`, `frontapp/src/lib/components/NotificationSettings.svelte`, `frontapp/src/lib/components/PWAInstallPromptModal.svelte` | `notification_handler.go` | `notification_repository.go` | `NOTIFICATION_TROUBLESHOOTING.md`, `frontapp/tests/unit/lib/components/NotificationSettings.svelte.spec.js`, `frontapp/tests/unit/lib/components/Sidebar.svelte.spec.js`, `frontapp/tests/unit/lib/stores/pwaInstallStore.svelte.spec.js` |
| MIC(行事委員会賞) | `frontapp/src/routes/dashboard/root/identify-mic/`, `frontapp/src/routes/dashboard/admin/vorting-mic/` | `mic_handler.go` | `mic_repository.go`, `mic.go` | `backapp/tests/handler/mic_handler_test.go`, `backapp/tests/repository/mic_repository_test.go`, `frontapp/tests/e2e/root-identify-mic.spec.js` |
| 競技ガイド資料 | `frontapp/src/routes/dashboard/root/competition-guidelines-upload/`, `frontapp/src/routes/dashboard/guide/` | `guide_document_handler.go`, `event_handler.go` | `guide_document_repository.go`, `guide_document.go` | `backapp/tests/handler/guide_document_handler_test.go`, `frontapp/tests/e2e/root-competition-guidelines-upload.spec.js` |
| 統計 | `frontapp/src/routes/dashboard/admin/manage-dashboard/` | `statistics_handler.go` | `class_repository.go`, `sport_repository.go`, `tournament_repository.go` | `backapp/tests/handler/statistics_handler_test.go` |
| システムバックアップ | 画面なし、root API | `system_handler.go` | DB dump、uploads dump | `backapp/tests/handler/system_handler_test.go`, `backapp/internal/handler/system_handler_dump_test.go` |
| WebSocket | トーナメント/進行状況表示画面 | `websocket_handler.go` | `backapp/internal/websocket/` | `backapp/tests/handler/websocket_handler_test.go`, `backapp/internal/websocket/*_test.go` |

### 性能・並行処理まわり

| 観点 | 主に見る場所 | メモ |
| --- | --- | --- |
| DB接続プール/共通最適化リソース | `backapp/internal/repository/db.go` | `SetMaxOpenConns` などのDB接続プール、`GlobalCache`、`GlobalSFGroup`、`GlobalAnts` の初期化 |
| Push通知送信の並行化 | `notification_handler.go`, `event_handler.go` | 通常通知とアンケート通知は `GlobalAnts` と `sync.WaitGroup` で購読ごとに並行送信する |
| 通知フィルタの並行化 | `notification_handler.go`, `event_handler.go` | ユーザーごとのフィルタ確認は `errgroup.SetLimit` でDB接続数を抑えながら並行化する |
| 統計APIのN+1確認 | `statistics_handler.go`, `class_repository.go`, `tournament_repository.go` | `GetClassScoreTrends` は `GetClassScoresByEvents`、進捗表示はsport名だけのJOIN取得を使う |
| トーナメント生成 | `all_tournament_handler.go`, `tournament_repository.go` | 競技ごとの読み取り/構造生成は `errgroup.SetLimit(4)`、DB保存は直列、match保存はbulk insertと一括 `next_match_id` 更新 |
| クラス進捗 | `class_handler.go`, `team_repository.go`, `tournament_repository.go` | 通常競技チーム、昼競技チーム、クラスメンバーは独立取得。試合/メンバー詳細は一括取得メソッドを優先する |

## DBとマイグレーション

| 場所 | 内容 |
| --- | --- |
| `backapp/db/migrations/000001_initial_schema.up.sql` | 初期スキーマ |
| `backapp/db/migrations/000001_initial_schema.down.sql` | 初期スキーマの取り消し |
| `backapp/db/migrations/000002_add_guide_documents.*.sql` | ガイド資料関連 |
| `backapp/db/migrations/000003_add_round_check_ins.*.sql` | ラウンドチェックイン関連 |
| `backapp/db/migrations/000005_add_duplicate_registration_threshold.*.sql` | 大会ごとの重複登録を許可するクラス人数上限 |
| `backapp/db/cleanup_score_logs_reason.sql` | スコアログ理由の整理用SQL |
| `backapp/db/ER図.pdf` | ER図 |

DB変更時は、既存のSQLを直接書き換えるのではなく、新しい番号の `*.up.sql` と `*.down.sql` を追加します。適用手順は `README.md` の「DBマイグレーション」を確認してください。

## テスト

| 場所/コマンド | 内容 |
| --- | --- |
| `backapp/tests/handler/` | API handler のテスト |
| `backapp/tests/repository/` | DB repository のテスト |
| `backapp/tests/middleware/` | middleware のテスト |
| `backapp/internal/**/**_test.go` | パッケージ内の近接テスト |
| `frontapp/tests/unit/` | Vitest の単体テスト |
| `frontapp/tests/e2e/` | Playwright のE2Eテスト |
| `cd backapp && go test ./...` | バックエンド全体テスト |
| `cd frontapp && npm run test` | フロント単体テスト |
| `cd frontapp && npm run test:e2e` | フロントE2Eテスト |

変更した機能に対応するテストがある場合は、まず該当テストを更新します。新しい画面やAPIを追加した場合は、最低限「成功系」と「権限/入力エラー系」のどちらをどこで担保するかを決めてください。

## 環境変数と起動設定

| 場所 | 内容 |
| --- | --- |
| `.env` | Docker/バックエンド共通の環境変数 |
| `frontapp/.env` | フロントエンド用の環境変数 |
| `frontapp/.env.sample` | フロントエンド環境変数のサンプル |
| `backapp/internal/config/config.go` | バックエンドが読む環境変数の一覧 |
| `frontapp/vite.config.js` | `BACKEND_URL` / `PUBLIC_BACKEND_URL` を使ったAPIプロキシ |
| `docker-compose.yml` | ローカル/検証用。Traefik、MySQL、Redis、migrate、frontapp、backapp |
| `docker-compose.production.yml` | 本番用。HTTPS、セキュリティヘッダー、Traefik設定 |

## 保守時のよくある探し方

### 画面の表示を直したい

1. URLから `frontapp/src/routes/.../+page.svelte` を探す。
2. 共通部品なら `frontapp/src/lib/components/` を探す。
3. データ取得がある画面は同じディレクトリの `+page.server.js` も見る。
4. APIレスポンスの形が合わない場合は `backapp/internal/router/router.go` からhandlerへ進む。

### APIを追加したい

1. `backapp/internal/router/router.go` にルートを追加する。
2. `backapp/internal/handler/` にhandlerを追加または既存handlerへメソッド追加する。
3. DB操作が必要なら `backapp/internal/repository/` に処理を追加する。
4. 入出力構造体が必要なら `backapp/internal/models/` に追加する。
5. 画面から呼ぶ場合は `frontapp/src/routes/...` または共通utilsから `fetch('/api/...')` する。
6. handler/repositoryテストを追加する。

### DBカラムを追加したい

1. `backapp/db/migrations/` に次の連番の `up/down` SQLを追加する。
2. 必要なら `backapp/internal/models/` の構造体を更新する。
3. `backapp/internal/repository/` のSELECT/INSERT/UPDATEを更新する。
4. handlerや画面で新項目を扱う。
5. repositoryテストと画面テストを更新する。

### 権限まわりを直したい

1. API制御は `backapp/internal/router/router.go` のグループとmiddlewareを確認する。
2. 判定ロジックは `backapp/internal/middleware/auth.go` を見る。
3. ユーザー/ロールDB処理は `backapp/internal/repository/user_repository.go` を見る。
4. フロントの表示出し分けは `dashboard` 配下のlayoutや各画面を確認する。
5. 仕様確認は `docs/user_permission_guide.md` を見る。

### 通知が届かない

1. `NOTIFICATION_TROUBLESHOOTING.md` を先に確認する。
2. フロントは `frontapp/src/service-worker.js`、`frontapp/src/lib/utils/push.js`、`frontapp/src/lib/stores/pushSubscriptionStore.js` を見る。
3. バックエンドは `notification_handler.go`、アンケート通知なら `event_handler.go`、購読/対象者取得は `notification_repository.go` を見る。
4. VAPID鍵は `.env` と `frontapp/.env` の両方を確認する。

### APIが遅い/N+1が疑わしい

1. 画面のURLから `router.go` とhandlerを特定する。
2. handler内でループ中にrepositoryを呼んでいないか確認する。
3. 複数IDを扱う読み取りは、まずrepositoryに `IN (...)` やJOINの一括取得を追加できるか検討する。
4. DB読み取りが独立していて一括化しにくい場合だけ、`errgroup.SetLimit` で並行化する。
5. DB書き込みはgoroutine化よりも、transaction、bulk insert、bulk updateを優先する。

### WebSocketを直したい

1. ルートは `backapp/internal/router/router.go` の `/api/ws/...` を確認する。
2. HTTPからWebSocketへの入口は `backapp/internal/handler/websocket_handler.go` を見る。
3. 接続管理とブロードキャストは `backapp/internal/websocket/` を見る。
4. フロント側はトーナメントや進行状況を表示する画面で `WebSocket` 利用箇所を検索する。

## 命名・配置の目安

| 追加したいもの | 置き場所 |
| --- | --- |
| 新しい画面 | `frontapp/src/routes/.../+page.svelte` |
| 画面専用のサーバー処理 | 画面と同じディレクトリの `+page.server.js` |
| 複数画面で使う部品 | `frontapp/src/lib/components/` |
| フロントの共有状態 | `frontapp/src/lib/stores/` |
| フロントの共有関数 | `frontapp/src/lib/utils/` |
| 新しいAPI | `backapp/internal/router/router.go` と `backapp/internal/handler/` |
| DBアクセス | `backapp/internal/repository/` |
| データ構造 | `backapp/internal/models/` |
| 認証・権限・CORS | `backapp/internal/middleware/` |
| WebSocket処理 | `backapp/internal/websocket/` |
| DBスキーマ変更 | `backapp/db/migrations/` |
| 運用・保守メモ | `docs/` |

## 注意点

- `frontapp/README.md` はSvelte雛形の説明が残っているため、実際の開発入口はルートの `README.md` を優先してください。
- Goのバージョンは `backapp/go.mod` が現在の正です。READMEや古いドキュメントと食い違う場合は `go.mod` を確認してください。
- APIの権限はhandler内だけでなく `router.go` のグループmiddlewareで決まっていることが多いです。
- アップロードファイルはバックエンドの `/uploads` に保存され、Dockerでは `uploads-data` volume に永続化されます。
- 本番DBでは `docker compose down` やvolume削除を不用意に実行しないでください。マイグレーション前はバックアップを取ります。
