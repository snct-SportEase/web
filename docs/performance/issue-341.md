# Issue #341: パフォーマンス改善

2026-10-05に実装・検証。変更前は `0732a39`。

## 最初の改善で確認した結果

同じmock backendとChromiumで初回表示からnetworkidleまでのリクエストを集計。
変更前のソースは一時ディレクトリへ展開し、集計用のmock endpointのみ追加した。
生データは [issue-341-measurements.json](issue-341-measurements.json)。

| 画面 | backend API総数: 前 → 後 | Active Event API: 前 → 後 | 大会一覧API: 前 → 後 |
| --- | --- | --- | --- |
| 学生マイページ | 13 → 11 | 3 → 1 | 1 → 0 |
| 学生得点一覧 | 7 → 5 | 3 → 1 | 1 → 0 |
| 学生クラス情報 | 8 → 6 | 3 → 1 | 1 → 0 |

変更後は通知subscription確認も1回観測されたため、API総数の減少は各画面2回。
大会関連の重複取得自体は各画面3回減少している。

sqlmockで同一ユーザーをTTL内に8回取得すると、User/RoleのSQLは16回から2回、
Active Event IDのSQLは1回となった。大会ID/詳細も8回の読み取りを各1回にまとめた。
同時miss、プロフィール変更、Role解除、大会変更、資料削除、競技解除をテストで確認した。

PlaywrightのnavigationMsはdev serverの変換とnetworkidleの500ms待機を含む単発の値で、
表示速度の改善率には使わない。本番の平均/p95、MySQL全クエリ数、DBプール待機は未計測。
DB接続数20/idle10は維持した。

## 実装

- User/Role、Active Event ID/詳細、大会競技一覧、資料一覧に5秒TTLのプロセス内キャッシュ。
- User/RoleキーはDB、ユーザー、大会ID、大会変更の世代を含む。
- プロフィール・Role更新は対象ユーザーを、大会更新・大会切替は大会世代を無効化。
- singleflightで同時missをまとめ、更新中の古い読み取りがキャッシュを再登録しないよう世代を確認。
- 各取得に独立したJSON snapshotを返し、呼び出し側のポインタ・slice変更を共有しない。
- 大会競技・資料更新、盤上競技作成、テスト大会restoreでも関連キャッシュを無効化。
- Dashboard hookで大会を一度取得し、SSRのlocalsとLayout dataを通して共有。
  browserでのみstoreへ渡し、SSRのグローバルstoreへユーザー間のデータを保存しない。
- storeは大会APIの情報だけで構築し、同時初期化を共有。更新操作と保守状態のpollは明示的に再取得。
- 得点一覧は得点APIの403を非表示として処理。クラス情報はclass-progressの大会IDを利用。
- 大会競技一覧と資料一覧のみ `private, max-age=0, must-revalidate` とETagを使用。
  304判定より先に認証・テスト隔離を確認し、個人データ・認証・通知・得点はno-storeを維持。
- 昼競技のテンプレート設定とタイピング結果Importを独立した遅延ロードcomponentへ分割。
  昼競技ページは94,691 → 74,264 bytes。学生・管理者グラフのChart.jsもdynamic import。
- frontend imageのビルドにnpm ciを使用し、実行imageにはproduction依存のみを含める。
  .dockerignoreでhostのnode_modules、生成物、環境ファイルの混入を防止。

キャッシュ無効化は同じbackendプロセス内で即反映する。複数プロセス・外部SQL更新では
最大5秒のTTL分の遅延がある。backendを複数replicaへ増やす場合は、特に権限失効について
Redis等での共有世代・無効化通知を先に実装する必要がある。

## Docker比較

Node 20 Alpine、同じpackage-lock、比較対象のDockerfileでproduction targetをビルド。

| Dockerfile | image bytes | 約MB |
| --- | ---: | ---: |
| 変更前 | 393,990,065 | 394 |
| 変更後 | 207,555,933 | 208 |

約47%削減。変更後imageの一時コンテナで `build/index.js` を起動しHTTP 200を確認。
Vite、Vitest、Playwright、Storybookが実行imageから解決できないことも確認した。

## 検証と本番での比較方法

- バックエンド全テストとrepositoryのrace detectorを通過。
- frontend server unit tests: 64件通過。
- 関連E2E: 得点非表示、競技情報、競技詳細設定、雨天切替、昼競技設定、タイピングImport、API回数を確認。
- frontend production buildとnpm ci済みbuilderでの変更ファイルlintを通過。

API回数は次で再現できる（利用中サービスと競合しないportを使用）。

```sh
cd frontapp
PLAYWRIGHT_HTML_OPEN=never PLAYWRIGHT_APP_PORT=5341 MOCK_BACKEND_PORT=8341 \
MOCK_BACKEND_URL=http://127.0.0.1:8341 npm run test:e2e -- dashboard-api-counts.spec.js
```

Dashboard応答には `Server-Timing: ssr;dur=...` を付ける。
本番ビルドで学生マイページのSSRとDashboard遷移を同じ端末・同じデータで複数回測り、
ブラウザーのNetwork/PerformanceでAPI回数と遷移時間を比較する。

backendで `PERFORMANCE_METRICS=true` を指定すると、30秒ごとにDB.Statsのopen/in_use/idle、
累積WaitCount/WaitDurationをログ出力する。区間差分を比較してから接続プールを調整する。
既存Ginログから指定4APIの成功/エラー別平均とp95を集計できる。

```sh
docker compose logs --no-log-prefix sportease-backapp | python3 scripts/summarize-api-latency.py
```

MySQLのクエリ総数は専用の負荷計測環境で同じユーザー・画面を実行し、
performance_schemaのdigest実行回数を区間差分で比較する。
本番計測と残る管理画面の追加分割は別途実施する。

## 追加の改善（2026-10-05）

`5fc028a` から追加で改善。生データは
[issue-341-phase2-measurements.json](issue-341-phase2-measurements.json)。

| 対象 | 前回 → 今回 | 確認方法 |
| --- | --- | --- |
| マイページのbackend API総数 | 11 → 10 | Chromium + mock backend |
| マイページの通知API | 2 → 1 | 同上 |
| クラス状況のSQL | 6 → 4 | handler mockとrepositoryのSQL経路 |
| 競技マスタ一覧のSQL（cache hit） | 3 → 0 | repository sqlmockとSQL経路 |
| 競技マスタ一覧のSQL（cache miss） | 4 → 1 | 同上、初期競技登録済みの場合 |

マイページは最初の変更前13回から10回になった。得点一覧5回、クラス情報6回も維持。
Dashboardトップでも大会APIは1回になり、プロフィール設定済みならクラス一覧APIを呼ばない。
Dashboardトップの総数は8回だったが、変更前の同条件計測がないため削減率は示さない。

クラス状況のSQL数は通常競技チームと割り当てチームがある場合で、認証・大会取得を除く。
`?view=summary` をマイページとクラス情報に指定し、割り当てチームとチーム所属学生の取得を省略。
学生一覧取得も `COUNT(*)` に替え、氏名・メール・割り当てを含む `members` は返さない。
handlerのテストで省略対象のrepositoryを呼ばず、人数・進捗を返すことを確認した。
メンバー情報が必要なDashboardトップは従来の取得を続ける。

競技マスタ一覧は、cache missの一覧取得で初期競技の不足も判断する。
競技作成・名称変更・テスト大会restoreで無効化し、5秒TTLとDB別の世代を使う。
初期競技作成中のcold readは直列化するが、更新前後の世代で結果を共有しない。

通知はSSRで50件を取得してサイドバーにIDだけを共有し、マイページ表示は従来どおり3件。
browser内でユーザー別の15秒snapshotと同時取得を共有する。
Pushは強制更新し、取得中に届いた場合は完了後に追加で取得する。
ユーザー変更・新しいSSRデータの到着後に古いレスポンスがバッジを上書きしないこともテストした。
認証と大会取得、Dashboardの独立した取得は並列に実行する。

プロセス内キャッシュはhit時にsingleflightへ入らず、世代参照にはRWMutexを使う。
12並列・各3回のmicrobenchmarkでは次の結果だった。

| cache hit | ns/opの中央値: 前 → 後 | bytes/op: 前 → 後 | allocs/op: 前 → 後 |
| --- | ---: | ---: | ---: |
| Active Event ID | 720.5 → 218.7 | 382 → 248 | 9 → 5 |
| User/Role snapshot | 1379 → 1358 | 822 → 672 | 24 → 20 |

Active Event IDのcache hit処理は約70%短縮。User/Roleは割り当てが減ったが、
処理時間の差は小さく速度改善とは判断しない。いずれもJSON復元を含むcache hitのみの計測で、
HTTP応答時間・本番p95の改善率ではない。

```sh
cd backapp
go test ./internal/repository -run '^$' -bench BenchmarkCachedReadHits \
  -benchmem -benchtime=200ms -count=3
```

追加変更後にバックエンド全テスト、repository/handlerのrace detector、frontend server unit tests
76件、関連E2E 14件、npm ci済みbuilderでのproduction buildと変更ファイルlintを通過。
E2EにはAPI回数、得点非表示、通知、競技作成・割り当て、クラスの競技割り当てを含む。
本番p95・DBプール待機は引き続き未計測。

## 敵対的テスト（2026-10-05）

最適化の境界を狙う28ケースを追加した。
同時実行の順序はGoのchannelとJavaScriptの待機promiseで制御し、sleepでタイミングを推測しない。

| テスト | 追加ケース | 検証する境界 |
| --- | ---: | --- |
| `read_cache_adversarial_test.go` | 4 | 3世代の読み取りを逆順に完了、DB/ユーザー/大会の分離、64並列の呼び出し側変更、JSON符号化失敗後の再試行 |
| `sport_defaults_cache_test.go` | 1 | 一部の初期競技登録だけ成功した後の再試行で重複登録しない |
| `class_progress_adversarial_test.go` | 9 | 未認証・クラス未所属・別大会・DB障害等で要約データの取得へ進まない |
| `hooks.server.adversarial.spec.js` | 6 | 認証/大会の通信・JSON障害、同時SSRを逆順に完了してもユーザーと大会が混ざらない |
| `notificationBadgeStore.adversarial.spec.js` | 8 | Pushの連続到着、503、ログアウト、ユーザー切替、JSON解析失敗、古い既読処理、破損した既読データ |

通知の2件は修正前にテストが失敗することを確認し、回帰テストとともに修正した。

- Pushによる追加取得の開始後に次のPushが届くと、追加取得を共有して最新通知を取りこぼしていた。
  開始後のPushを記録し、完了後にもう一度取得する。同時に届いたPushはまとめる。
- 前ユーザーの既読処理が遅れて実行されると、現在のユーザーの通知バッジも消していた。
  既読保存は対象ユーザーに行い、現在のバッジは同じユーザーの場合だけ変更する。

キャッシュのテストについても、作業ツリーを変更しないGo overlayで
世代をキーから除く変更とDB識別子を除く変更をそれぞれ適用し、追加テストが失敗することを確認した。
この一時変更はコミットしていない。

競合を含む主要ケースは次を20回繰り返してrace detectorを通過した。

```sh
cd backapp
go test -race -count=20 ./internal/repository ./tests/handler \
  -run 'TestCachedReadOverlappingRevisions|TestCachedReadConcurrentCallerMutations|TestSportDefaultsPartialFailure|TestClassProgressSummaryRejects'
```

変更後のバックエンド全テスト、repository/handlerのrace detector、frontend server unit tests
90件、通知・得点非表示・API回数のE2E 8件、production build、変更ファイルlintを通過。
