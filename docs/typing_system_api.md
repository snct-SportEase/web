# SportEase提供・独自タイピングシステム連携API契約

| 項目 | 内容 |
| --- | --- |
| 文書状態 | v1確定 |
| APIバージョン | v1 |
| 最終更新 | 2026-08-02 |
| 提供側 | SportEase |
| 利用側 | 独自タイピングシステム |

SportEase (`https://nitsche-gyouji.com`) から、独自タイピングシステム (`https://typing.nitsche-gyouji.com`) が競技実行に必要な大会・出場情報を取得するAPIの必須契約である。主催者が決定するコア機能とコンペ参加開発者へ任せる範囲は、[共通仕様の範囲](./typing_system_core_scope.md)を参照する。

## 1. 接続契約

| 項目 | 契約 |
| --- | --- |
| ベースURL | `https://nitsche-gyouji.com/api/integrations/typing/v1` |
| 通信方向 | タイピングシステムのバックエンド → SportEase |
| プロトコル | HTTPSのみ |
| データ形式 | UTF-8のJSON、フィールド名は`snake_case` |
| 認証 | `Authorization: Bearer <token>` |
| 日時 | UTCのRFC 3339形式。例: `2026-09-01T03:00:00Z` |

- APIはブラウザから直接呼び出さず、サーバー間通信に限定する。
- このAPIのトークンは「確定スコアAPI」のトークンとは別に発行する。
- トークンをブラウザ、Gitリポジトリ、Dockerイメージ、ログへ含めない。
- 成功レスポンスには`api_version`、`generated_at`、`snapshot_version`を含める。

## 2. エンドポイント

### 2.1 操作対象の大会・競技

```http
GET /api/integrations/typing/v1/events/active
Authorization: Bearer <token>
```

```json
{
  "api_version": "v1",
  "generated_at": "2026-08-31T23:00:00Z",
  "snapshot_version": 7,
  "events": [
    {
      "event_id": 12,
      "name": "秋季スポーツ大会",
      "year": 2026,
      "season": "autumn",
      "status": "active",
      "starts_at": "2026-08-31T15:00:00Z",
      "ends_at": "2026-09-01T15:00:00Z",
      "sports": [
        {
          "sport_id": 8,
          "name": "タイピング"
        }
      ]
    }
  ]
}
```

- タイピング競技を含み、運営対象となる大会だけを返す。
- 対象がない場合は`events: []`を返す。

### 2.2 トーナメント一覧

```http
GET /api/integrations/typing/v1/events/{event_id}/sports/{sport_id}/tournaments
Authorization: Bearer <token>
```

```json
{
  "api_version": "v1",
  "generated_at": "2026-09-01T00:00:00Z",
  "snapshot_version": 11,
  "event_id": 12,
  "sport_id": 8,
  "tournaments": [
    {
      "tournament_id": 31,
      "name": "3年生 タイピング",
      "status": "ready",
      "match_count": 3,
      "team_count": 6,
      "updated_at": "2026-09-01T00:00:00Z"
    }
  ]
}
```

- ページングは使用せず、指定大会・競技のタイピング用トーナメントを全件返す。
- `match_count`は3、`team_count`は6でなければならない。不足時もデータを返し、詳細の`validation_errors`で通知する。

### 2.3 トーナメント競技スナップショット

```http
GET /api/integrations/typing/v1/tournaments/{tournament_id}
Authorization: Bearer <token>
```

```json
{
  "api_version": "v1",
  "generated_at": "2026-09-01T00:05:00Z",
  "snapshot_version": 12,
  "event": {
    "event_id": 12,
    "name": "秋季スポーツ大会",
    "year": 2026,
    "season": "autumn",
    "status": "active",
    "starts_at": "2026-08-31T15:00:00Z",
    "ends_at": "2026-09-01T15:00:00Z"
  },
  "sport": {
    "sport_id": 8,
    "name": "タイピング"
  },
  "tournament": {
    "tournament_id": 31,
    "name": "3年生 タイピング",
    "status": "ready"
  },
  "matches": [
    {
      "match_id": 205,
      "match_number": 1,
      "status": "scheduled",
      "scheduled_at": "2026-09-01T03:00:00Z",
      "duration_seconds": 180,
      "participants": [
        {
          "lane_number": 1,
          "team_id": 71,
          "team_name": "3年1組",
          "class_id": 301,
          "class_name": "3-1",
          "player_id": "e4eaa4c4-48d0-4ce0-8c60-31c6b5531d65",
          "player_display_name": "選手A",
          "entry_order": 1
        }
      ]
    }
  ],
  "validation_errors": []
}
```

例では`matches`と`participants`を1件だけ記載しているが、正常なスナップショットは3試合を持ち、各試合に6件の`participants`を持つ。

## 3. データモデルと検証規則

### 3.1 タイピング専用試合

- 各タイピング用トーナメントに、第1〜第3試合の3レコードを作る。
- SportEaseが各試合へ一意な整数の`match_id`を発行する。
- 6チームから1名ずつ、合計6名を各試合へ割り当てる。
- 各チームの`entry_order` 1〜3を第1〜第3試合へ対応させる。
- `lane_number`は各試合で1〜6の一意な整数とし、3試合を通して同じチームは同じレーンを使用する。
- `duration_seconds`は同じトーナメントの3試合で共通とする。

### 3.2 必須ID

`event_id`、`sport_id`、`tournament_id`、`match_id`、`team_id`、`class_id`は整数とする。`player_id`はSportEaseが発行した文字列をそのまま使用する。タイピングシステムはこれらを独自IDへ置き換えず、結果APIへ同じ値を返す。

### 3.3 snapshot_version

- `snapshot_version`は、対象レスポンスの競技実行に関係するデータが変わるたび1増加する単調増加整数とする。
- 同じ内容に対する再取得では同じ値を返す。
- タイピングシステムは試合準備時に取得したトーナメント詳細と`snapshot_version`を保存する。
- 試合開始直前に再取得し、保存済みと異なる場合は運営者へ変更内容を示して再確認させる。
- 試合開始後は、当該試行の出場者、レーン、競技時間を保存済みスナップショットから変更しない。

### 3.4 validation_errors

開始できない構成上の問題を次の形式で返す。

```json
{
  "code": "participant_count_mismatch",
  "message": "Match 205 must have exactly 6 participants.",
  "path": "matches[0].participants"
}
```

最低限、次を検証する。

- トーナメントが3試合で構成されている。
- 参加チームが6チームである。
- 各チームに本登録済み選手が3名おり、`entry_order` 1〜3が重複なく設定されている。
- 各試合に異なる6チーム・6選手・6レーンが割り当てられている。
- `match_number`と`entry_order`が一致する。
- 競技時間が設定され、3試合で一致する。

`validation_errors`が1件以上ある場合、タイピングシステムは正式試合を開始してはならない。メールアドレス、ロール、ログイン情報はレスポンスへ含めない。

## 4. 出場順の設定（SportEase管理画面用）

この操作は外部連携APIではない。SportEaseの`admin`または`root`セッションと通常のCSRF保護を使用する。

```http
PUT /api/admin/typing/teams/{team_id}/entry-order
Content-Type: application/json

{
  "player_ids": [
    "player-id-for-match-1",
    "player-id-for-match-2",
    "player-id-for-match-3"
  ]
}
```

`player_ids`は、そのチームで本登録済みの3選手と完全に一致しなければならない。配列順が`entry_order` 1、2、3となる。

## 5. エラー形式

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
| `404` | `event_not_found`、`sport_not_found`、`tournament_not_found` | 対象が存在しない |
| `409` | `entry_order_mismatch` | 管理画面で指定した選手と本登録選手が一致しない |
| `429` | `rate_limit_exceeded` | レート制限超過 |
| `500` | `internal_error` | SportEase内部エラー |
| `503` | `service_unavailable` | 一時的にデータを取得できない、または認証設定がない |

`message`は人間向けであり、処理分岐には`code`を使用する。内部スタックトレースや秘密情報を返してはならない。

## 6. 確定スコアの逆方向連携

このAPIは「SportEase → タイピングシステム」の大会・出場情報を提供する。逆方向の確定スコアは、タイピングシステム側がAPIを提供し、SportEaseが取得する。

詳細は[タイピングシステム確定スコアAPI契約](./typing_system_score_api.md)を参照する。タイピング運営画面で運営者が確定した1試合6名分だけを`confirmed`として渡し、`provisional`や部分確定を渡してはならない。

## 7. 実装上の注意

この文書はv1の確定契約である。既存のSportEase実装が旧エンドポイント、`X-API-Key`、通常競技用の対戦モデルを使用している場合は、本番接続前にこの契約へ更新する必要がある。

## 8. 変更履歴

| 日付 | 内容 |
| --- | --- |
| 2026-08-02 | v1 API契約を確定。Bearer認証、トーナメント一覧・詳細、3試合×6名のスナップショット、共通エラー形式を決定 |
| 2026-08-02 | 初版。SportEaseから大会・競技情報を取得する方向と出場順設定を記録 |
