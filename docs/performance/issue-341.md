# Issue #341: パフォーマンス改善

2026-10-05に実装・検証。変更前は `0732a39`。

## ローカルで確認した結果

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
