# タイピングシステム確定スコアAPI契約

| 項目 | 内容 |
| --- | --- |
| 文書状態 | v1確定 |
| APIバージョン | v1 |
| 最終更新 | 2026-08-02 |
| 提供側 | 独自タイピングシステム |
| 利用側 | SportEase |

この文書は、独自タイピングシステムからSportEaseへ確定スコアを連携するAPIの必須契約である。画面構成や内部実装など、参加開発者へ任せる範囲は[共通仕様の範囲](./typing_system_core_scope.md)を正とする。

## 1. 接続契約

| 項目 | 契約 |
| --- | --- |
| ベースURL | `https://typing.nitsche-gyouji.com/api/integrations/sportease/v1` |
| 通信方向 | SportEaseバックエンド → タイピングシステム |
| プロトコル | HTTPSのみ |
| データ形式 | UTF-8のJSON、フィールド名は`snake_case` |
| 認証 | `Authorization: Bearer <token>` |
| 日時 | UTCのRFC 3339形式。例: `2026-09-01T03:04:15Z` |

- APIはブラウザから直接呼び出さず、サーバー間通信に限定する。
- このAPIのトークンは「SportEase提供API」のトークンとは別に発行する。
- トークンをブラウザ、Gitリポジトリ、Dockerイメージ、ログへ含めない。
- 成功レスポンスには`api_version`、`generated_at`、`snapshot_version`を含める。

## 2. 責務と取得方式

- タイピングシステムが確定スコア取得APIを提供し、SportEaseが定期的に取得する。
- 正式結果の正本、3名分のチーム合計、大会順位はSportEaseで管理する。
- 運営者がタイピング運営画面で「スコア確定」を実行した結果だけを`confirmed`として公開する。
- 1試合6名分を1つのバッチとして確定し、部分的な確定結果は公開しない。
- SportEaseはトーナメント単位で結果を毎回全件取得する。ページングおよび差分取得は使用しない。
- 1トーナメントは3試合、各試合6名のため、現在有効な結果は最大18件とする。
- SportEaseは`result_id`と`revision`を使って冪等に取り込む。

## 3. 試合・試行・結果の識別子

| フィールド | 単位 | 規則 |
| --- | --- | --- |
| `match_id` | SportEase上の試合 | SportEaseが発行する整数。同じ第1〜第3試合では再試合後も同じ値を使う |
| `attempt_id` | 1回の6名同時競技 | タイピングシステムが発行するUUID。再試合では新しい値を発行する |
| `confirmation_batch_id` | 6名分の一括確定 | タイピングシステムが発行するUUID。同時に確定した6件で共通とする |
| `result_id` | 1選手・1試行の結果 | タイピングシステムが発行するUUID。発行後は変更・再利用しない |
| `revision` | 1件の結果の公開版 | 1から開始し、公開後に状態が変わるたび1増加させる |

- 各タイピング用トーナメントに第1〜第3試合の3レコードを作る。
- 各試合には6チームから1名ずつ、合計6名を割り当てる。
- 各チームの`entry_order` 1〜3を第1〜第3試合へ対応させる。
- 競技の再実施では`match_id`を維持し、`attempt_id`、`confirmation_batch_id`、6件の`result_id`を新しく発行する。
- 確定済みの数値を上書き訂正してはならない。誤った試行を無効化し、必要なら再試合を新しい試行として確定する。

## 4. エンドポイント

### 4.1 ヘルスチェック

```http
GET /api/integrations/sportease/v1/health
Authorization: Bearer <token>
```

```json
{
  "api_version": "v1",
  "generated_at": "2026-09-01T03:00:00Z",
  "status": "ok"
}
```

認証、アプリケーション、データベースが正常で、結果取得を受け付けられる場合に`200 OK`を返す。

### 4.2 トーナメント確定結果の全件取得

```http
GET /api/integrations/sportease/v1/tournaments/{tournament_id}/results
Authorization: Bearer <token>
```

トーナメントの第1〜第3試合について、現在SportEaseが採用すべき最新の試行を返す。`provisional`の試行はレスポンスに含めない。確定バッチが6件揃っていない試行も含めない。

```json
{
  "api_version": "v1",
  "generated_at": "2026-09-01T03:05:00Z",
  "snapshot_version": 14,
  "tournament_id": 31,
  "matches": [
    {
      "match_id": 205,
      "match_number": 1,
      "attempt_id": "1eb21ee4-00bd-434f-a997-e618ceba9f4a",
      "confirmation_batch_id": "69fdc70f-ec4f-43fb-b9a4-f577ae898808",
      "status": "confirmed",
      "supersedes_attempt_id": null,
      "confirmed_at": "2026-09-01T03:04:15Z",
      "updated_at": "2026-09-01T03:04:15Z",
      "results": [
        {
          "result_id": "7b391f04-8577-49d7-96da-036dfa3b32c3",
          "revision": 1,
          "event_id": 12,
          "sport_id": 8,
          "tournament_id": 31,
          "match_id": 205,
          "attempt_id": "1eb21ee4-00bd-434f-a997-e618ceba9f4a",
          "confirmation_batch_id": "69fdc70f-ec4f-43fb-b9a4-f577ae898808",
          "team_id": 71,
          "player_id": "e4eaa4c4-48d0-4ce0-8c60-31c6b5531d65",
          "entry_order": 1,
          "lane_number": 1,
          "metrics": {
            "correct_types": 421,
            "incorrect_types": 18,
            "wpm": 140.3333333333,
            "accuracy": 0.9589977221,
            "duration_seconds": 180
          },
          "score": {
            "total": 123,
            "formula_version": "wpm-accuracy-cubed-v1"
          },
          "status": "confirmed",
          "started_at": "2026-09-01T03:00:00Z",
          "ended_at": "2026-09-01T03:03:00Z",
          "confirmed_at": "2026-09-01T03:04:15Z",
          "updated_at": "2026-09-01T03:04:15Z"
        }
      ]
    }
  ]
}
```

例では`results`を1件だけ記載しているが、`status`が`confirmed`の試合では必ず6件を返す。

### 4.3 結果1件の再取得

```http
GET /api/integrations/sportease/v1/results/{result_id}
Authorization: Bearer <token>
```

レスポンスの`result`には4.2の結果要素と同じ形式を返す。過去の試行が`invalidated`または`superseded`になった後も、その`result_id`で監査用に再取得できなければならない。

## 5. 結果フィールド

| フィールド | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| `result_id` | UUID文字列 | 必須 | 1選手・1試行の不変ID |
| `revision` | 1以上の整数 | 必須 | 同じ`result_id`の公開版 |
| `event_id` | 整数 | 必須 | SportEaseの大会ID |
| `sport_id` | 整数 | 必須 | SportEaseの競技ID |
| `tournament_id` | 整数 | 必須 | SportEaseのタイピング用トーナメントID |
| `match_id` | 整数 | 必須 | SportEaseのタイピング専用試合ID |
| `attempt_id` | UUID文字列 | 必須 | 6名共通の試行ID |
| `confirmation_batch_id` | UUID文字列 | 必須 | 6名共通の一括確定ID |
| `team_id` | 整数 | 必須 | SportEaseのチームID |
| `player_id` | 文字列 | 必須 | SportEaseの選手ID |
| `entry_order` | 1〜3の整数 | 必須 | チーム内の出場順。試合番号と一致する |
| `lane_number` | 1〜6の整数 | 必須 | 当該試合のレーン番号 |
| `metrics.correct_types` | 0以上の整数 | 必須 | 正タイプ数 |
| `metrics.incorrect_types` | 0以上の整数 | 必須 | 誤タイプ数 |
| `metrics.wpm` | 0以上の数 | 必須 | `correct_types / duration_seconds * 60` |
| `metrics.accuracy` | 0〜1の数 | 必須 | `correct_types / (correct_types + incorrect_types)`。両方0なら0 |
| `metrics.duration_seconds` | 0より大きい数 | 必須 | 設定された競技時間。6名共通 |
| `score.total` | 0以上の整数 | 必須 | `floor(wpm * accuracy^3)` |
| `score.formula_version` | 文字列 | 必須 | `wpm-accuracy-cubed-v1` |
| `status` | 列挙文字列 | 必須 | `confirmed`、`invalidated`、`disqualified`、`superseded`のいずれか |
| `started_at` | RFC 3339日時 | 必須 | 試合開始時刻 |
| `ended_at` | RFC 3339日時 | 必須 | 試合終了時刻 |
| `confirmed_at` | RFC 3339日時 | 必須 | 運営者が確定した時刻 |
| `updated_at` | RFC 3339日時 | 必須 | 当該revisionの更新時刻 |

選手名、クラス名、完了文字数、完了問題数は連携に必須ではない。表示用の所属情報はSportEaseのIDから取得し、スコアの正本は正タイプ数、誤タイプ数、競技時間とする。計算の詳細は[スコア計算仕様](./typing_system_scoring_spec.md)を参照する。

## 6. 状態と再試合

### 6.1 状態

| 状態 | APIへの公開 | 意味 |
| --- | --- | --- |
| `provisional` | 非公開 | 競技終了後、運営者による確定前 |
| `confirmed` | 公開 | 運営者が6名分を一括確定済み |
| `invalidated` | 公開 | 確定後、試合全体を無効化した |
| `disqualified` | 公開 | 確定後、対象選手を失格とした |
| `superseded` | 詳細取得のみ | 再試合の確定結果に置き換えられた過去の結果 |

```text
競技終了
  → provisionalとして保存
  → 運営者が6名分を一括確定
  → confirmedとして公開
  → 必要な場合はinvalidatedまたはdisqualified
  → 再試合を確定した場合、以前の試行をsuperseded
```

- 自動処理だけで`confirmed`へ変更してはならない。
- 設定時間より前に終了した試合は`confirmed`にせず、無効な試行として再試合する。
- 試合全体の`status`は6件すべてが`confirmed`なら`confirmed`、6件すべてが`invalidated`なら`invalidated`、1件以上が`disqualified`なら`disqualified`とする。
- 再試合を確定した場合、4.2では新しい試行だけを返し、`supersedes_attempt_id`に直前の正式試行IDを設定する。SportEaseは直前の試行を正式集計から外す。
- 過去の試行は4.3で取得可能な状態を維持する。

### 6.2 revisionの取り込み

- 同じ`result_id`の初回公開は`revision: 1`とする。
- 状態または公開フィールドが変わる場合は`revision`を1増加させる。
- SportEaseに未登録の`result_id`は登録する。
- 登録済みの場合、受信した`revision`が保存済みより大きいときだけ更新する。
- 同じ`revision`は処理済みとして成功扱いにし、小さい`revision`は無視する。
- `snapshot_version`はトーナメントの公開結果集合が変わるたび1増加する単調増加整数とする。同じ値ならSportEaseは取り込みを省略できる。

## 7. 整合性検証

SportEaseは取り込み時に次を検証し、不一致があれば正式集計へ反映せずエラー記録を残す。

- `event_id`、`sport_id`、`tournament_id`、`match_id`がSportEaseの割り当てと一致する。
- `team_id`、`player_id`、`entry_order`、`lane_number`が試合開始時のスナップショットと一致する。
- `confirmed`の試合は同じ`attempt_id`と`confirmation_batch_id`を持つ6件で構成される。
- 6件の`match_id`と`duration_seconds`が共通で、チーム、選手、レーンが重複しない。
- WPM、正確率、総合スコアが元データから再計算した値と一致する。
- `formula_version`がSportEaseの対応バージョンである。

小数値の文字列表現の差で判定が変わらないよう、正式判定と同点比較では[スコア計算仕様](./typing_system_scoring_spec.md)に記載した整数による同値式を使用する。

## 8. エラー形式

エラー時は次の共通形式を返す。

```json
{
  "error": {
    "code": "tournament_not_found",
    "message": "The requested tournament was not found.",
    "request_id": "01J6Q7N6N4GF3R5VTVJ1MWP8DB"
  }
}
```

| HTTP状態 | `code`例 | 意味 |
| --- | --- | --- |
| `400` | `invalid_request` | IDまたはリクエスト形式が不正 |
| `401` | `unauthorized` | Bearerトークンがない、または不正 |
| `404` | `tournament_not_found`、`result_not_found` | 対象が存在しない |
| `429` | `rate_limit_exceeded` | レート制限超過 |
| `500` | `internal_error` | タイピングシステム内部エラー |
| `503` | `service_unavailable` | 一時的に結果を取得できない |

`message`は人間向けであり、処理分岐には`code`を使用する。内部スタックトレースや秘密情報を返してはならない。

## 9. 監査と保持

- 確定、無効化、失格、再試合、差し替えについて、操作日時、操作主体、対象ID、変更前後の状態を記録する。
- SportEase停止中も確定結果を失わず、復旧後に同じAPIから取得できるよう保持する。
- APIアクセスには`request_id`、時刻、対象、HTTP状態を記録し、Bearerトークンは記録しない。

## 10. 変更履歴

| 日付 | 内容 |
| --- | --- |
| 2026-08-02 | v1 API契約を確定。Bearer認証、全件取得、最大18件、識別子、revision、再試合、状態、エラー形式を決定 |
| 2026-08-02 | ゼロ入力時は0、途中終了した試合は確定不可、同点比較順を決定 |
| 2026-08-02 | 初版。試合ID、6名一括確定、結果リソース、スコア計算を反映 |
