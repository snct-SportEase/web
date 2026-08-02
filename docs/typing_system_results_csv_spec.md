# 独自タイピングシステム 確定結果CSV仕様

| 項目 | 内容 |
| --- | --- |
| 文書状態 | コア採用決定・列定義v1案 |
| スキーマバージョン | `typing-results-v1` |
| 最終更新 | 2026-08-02 |
| 出力側 | 独自タイピングシステム |
| 取込側 | SportEase管理画面 |

## 1. 目的と運用

独自タイピングシステムは、運営者が確定した競技結果をCSVファイルとして出力する。運営者はSportEaseで取込先の大会・タイピング用トーナメントを選択し、CSVをアップロードして検証結果を確認した後、正式結果へ反映する。

- CSVはトーナメント単位で出力する。
- 正常終了時は3試合×6名の18行を含める。
- 各試合は6名全員が確定済みの場合だけ含める。
- `provisional`の結果や6名未満の部分確定を含めない。
- CSV出力後も、タイピングシステムだけで結果を閲覧・再出力できること。

## 2. ファイル形式

| 項目 | 仕様 |
| --- | --- |
| 文字コード | UTF-8。BOM付きで出力する |
| 区切り文字 | カンマ `,` |
| 改行 | CRLF |
| ヘッダー | 1行目に必須。列名は4章のとおり |
| クォート | RFC 4180に従い、カンマ、改行、ダブルクォートを含む値をダブルクォートで囲む |
| 小数点 | ピリオド `.` |
| 日時 | UTCのRFC 3339形式。例: `2026-09-01T03:04:15Z` |
| 空値 | 任意列だけ空文字を許可。必須列は空にしない |

`wpm`と`accuracy`は、元データから求めた値を小数点以下10桁へ四捨五入して出力する。SportEaseはこれらの表示値から総合スコアを計算せず、`correct_types`、`incorrect_types`、`duration_seconds`から再計算する。

ファイル名は次を推奨する。

```text
typing-results_{tournament_code}_{exported_at}.csv
```

## 3. 識別子と重複防止

| 識別子 | 規則 |
| --- | --- |
| `export_id` | 1回のCSV出力に対して発行するUUID。全行で同じ値 |
| `attempt_id` | 1試合6名の1回の競技に対して発行するUUID。6行で共通 |
| `confirmation_batch_id` | 6名分の一括確定に対して発行するUUID。6行で共通 |
| `result_id` | 1選手・1試行に対して発行するUUID。再利用しない |

- SportEaseは`result_id`を取込済み結果の重複判定に使用する。
- 同じ`result_id`、同じ内容は取込済みとして成功扱いにする。
- 同じ`result_id`で内容が異なる場合はファイル全体をエラーにする。
- 再試合では新しい`attempt_id`、`confirmation_batch_id`、6件の`result_id`を発行する。
- 再試合結果は`supersedes_attempt_id`で直前の正式試行を指定する。

## 4. 列定義

CSVの列順は次のとおりとする。

| # | 列名 | 型 | 必須 | 説明 |
| --- | --- | --- | --- | --- |
| 1 | `schema_version` | 文字列 | 必須 | `typing-results-v1`固定 |
| 2 | `export_id` | UUID | 必須 | CSV出力ID |
| 3 | `exported_at` | 日時 | 必須 | CSV出力日時。全行共通 |
| 4 | `event_code` | 文字列 | 必須 | 大会を照合する運営コード |
| 5 | `event_name` | 文字列 | 必須 | 確認表示用の大会名 |
| 6 | `tournament_code` | 文字列 | 必須 | トーナメントを照合する運営コード |
| 7 | `tournament_name` | 文字列 | 必須 | 確認表示用のトーナメント名 |
| 8 | `match_number` | 整数 | 必須 | 1〜3 |
| 9 | `attempt_id` | UUID | 必須 | 試行ID。同じ試合の6行で共通 |
| 10 | `supersedes_attempt_id` | UUID | 任意 | 再試合で置き換える直前の試行ID |
| 11 | `confirmation_batch_id` | UUID | 必須 | 一括確定ID。同じ試合の6行で共通 |
| 12 | `result_id` | UUID | 必須 | 選手結果ID |
| 13 | `lane_number` | 整数 | 必須 | 1〜6。試合内で一意 |
| 14 | `entry_order` | 整数 | 必須 | 1〜3。`match_number`と一致 |
| 15 | `class_code` | 文字列 | 必須 | SportEaseとの照合用クラスコード |
| 16 | `class_name` | 文字列 | 必須 | 確認表示用クラス名 |
| 17 | `team_code` | 文字列 | 必須 | SportEaseとの照合用チームコード |
| 18 | `team_name` | 文字列 | 必須 | 確認表示用チーム名 |
| 19 | `player_code` | 文字列 | 必須 | SportEaseとの照合用選手コード |
| 20 | `player_name` | 文字列 | 必須 | 確認表示用選手名 |
| 21 | `problem_set_id` | 文字列 | 必須 | 使用した問題セットID |
| 22 | `problem_set_version` | 1以上の整数 | 必須 | 使用した問題セットの版 |
| 23 | `correct_types` | 0以上の整数 | 必須 | 正タイプ数 |
| 24 | `incorrect_types` | 0以上の整数 | 必須 | 誤タイプ数 |
| 25 | `duration_seconds` | 0より大きい整数 | 必須 | 設定された競技時間 |
| 26 | `wpm` | 0以上の数 | 必須 | 小数点以下10桁の表示値 |
| 27 | `accuracy` | 0〜1の数 | 必須 | 小数点以下10桁の表示値 |
| 28 | `score` | 0以上の整数 | 必須 | 総合スコア |
| 29 | `formula_version` | 文字列 | 必須 | `wpm-accuracy-cubed-v1` |
| 30 | `status` | 列挙文字列 | 必須 | `confirmed`、`invalidated`、`disqualified` |
| 31 | `started_at` | 日時 | 必須 | 試合開始時刻 |
| 32 | `ended_at` | 日時 | 必須 | 試合終了時刻 |
| 33 | `confirmed_at` | 日時 | 必須 | 運営者による確定日時 |

SportEase内部の数値IDをCSV必須項目にしない。`event_code`、`tournament_code`、`class_code`、`team_code`、`player_code`は、SportEaseの登録値と一致する運営コードをタイピングシステムへ事前登録する。SportEaseから出場者テンプレートCSVを出力して取り込む機能を追加してもよいが、タイピングシステムは手動登録だけでも競技を完結できなければならない。

### ヘッダー

```csv
schema_version,export_id,exported_at,event_code,event_name,tournament_code,tournament_name,match_number,attempt_id,supersedes_attempt_id,confirmation_batch_id,result_id,lane_number,entry_order,class_code,class_name,team_code,team_name,player_code,player_name,problem_set_id,problem_set_version,correct_types,incorrect_types,duration_seconds,wpm,accuracy,score,formula_version,status,started_at,ended_at,confirmed_at
```

### データ行の例

```csv
typing-results-v1,4ef87d60-2e74-477a-9c16-a93423d04c20,2026-09-01T03:10:00Z,2026-autumn,秋季スポーツ大会,grade-3,3年生タイピング,1,1eb21ee4-00bd-434f-a997-e618ceba9f4a,,69fdc70f-ec4f-43fb-b9a4-f577ae898808,7b391f04-8577-49d7-96da-036dfa3b32c3,1,1,3-1,3年1組,3-1-typing,3年1組,student-30101,選手A,set-2026-a,1,421,18,180,140.3333333333,0.9589977221,123,wpm-accuracy-cubed-v1,confirmed,2026-09-01T03:00:00Z,2026-09-01T03:03:00Z,2026-09-01T03:04:15Z
```

## 5. SportEase取込検証

SportEaseはアップロード直後に検証結果を表示し、エラーが0件の場合だけ「正式結果へ反映」を許可する。

最低限、次を検証する。

- ヘッダー、スキーマバージョン、文字コード、型、必須値が正しい。
- 全行の`export_id`、`exported_at`、大会、トーナメントが一致する。
- 試合ごとに6行あり、レーン、チーム、選手が重複していない。
- `match_number`と`entry_order`が一致する。
- 取込先の大会、トーナメント、クラス、チーム、選手コードと一致する。
- 同じ試合の`attempt_id`、`confirmation_batch_id`、競技時間、問題セットが一致する。
- 正タイプ数、誤タイプ数、競技時間からWPM、正確率、総合スコアを再計算して一致する。
- `formula_version`がSportEaseの対応バージョンである。
- `confirmed`の部分確定や、置換関係が不正な再試合結果がない。
- `wpm`と`accuracy`が規定の10桁表示値と一致する。

取込確認画面には、取込対象の3試合、18名、チーム別3名合計、警告、エラーを表示する。正式反映前に運営者が内容を確認できなければならない。

## 6. 原子性と訂正

- 1つのCSVは全行成功または全行失敗とし、部分的に登録しない。
- 正式反映後に同じファイルを再度取り込んでも結果を増やさない。
- 確定済み結果の数値を同じ`result_id`で上書きしない。
- 誤った試行は無効化し、再試合結果を新しいIDで出力する。
- CSV取込と正式反映について、操作日時、操作主体、`export_id`、ファイル名、成功・失敗をSportEaseへ記録する。

## 7. CSVの安全性

- セル値の先頭が`=`、`+`、`-`、`@`の場合、表計算ソフトで数式として実行されないよう無害化する。
- CSVへメールアドレス、ログイン情報、認証情報、キー入力履歴を含めない。
- SportEaseはファイルサイズ、行数、列数、文字数を制限し、サーバー側で再検証する。
- CSVの値をHTMLへ表示するときはエスケープする。

## 8. 未決事項

| ID | 未決事項 | 決定が必要な理由 |
| --- | --- | --- |
| CSV-001 | 大会・クラス・チーム・選手の運営コード発行方法 | 両システムで確実に照合するため |
| CSV-002 | 3試合すべて確定後だけ出力するか、試合ごとの中間CSVも許可するか | 当日の運営手順に影響するため |
| CSV-003 | 無効・失格時のSportEase得点処理 | 正式順位に影響するため |

## 9. 変更履歴

| 日付 | 内容 |
| --- | --- |
| 2026-08-02 | 確定結果CSVをSportEaseへの結果登録方式として作成 |
