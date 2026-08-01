# 独自タイピングシステム連携API

SportEase (`https://nitsche-gyouji.com`) から、独自タイピングシステム (`https://typing.nitsche-gyouji.com`) が競技実行に必要な情報を取得するためのAPIです。

## 接続方式

- 通信方向: タイピングシステムのバックエンド → SportEase
- ベースURL: `https://nitsche-gyouji.com/api/integrations/typing/v1`
- 認証ヘッダー: `X-API-Key: <TYPING_API_KEY>`
- 制限: 送信元IPごとに1分120リクエスト
- キャッシュ: 無効（各レスポンスに取得時刻 `generated_at` を含む）

APIキーをブラウザのJavaScript、Gitリポジトリ、Dockerイメージへ埋め込まないでください。タイピングシステムのサーバー側環境変数へ保存します。このAPIはサーバー間利用を前提とするため、ブラウザからのクロスオリジン呼び出しは許可していません。

SportEase側では、例えば次のように共有鍵を作成してルート `.env` の `TYPING_API_KEY` に設定します。

```bash
openssl rand -hex 32
```

未設定時は誤公開を避けるため、連携APIは `503 Service Unavailable` を返します。

## エンドポイント

### 操作対象の大会と競技一覧

```http
GET /api/integrations/typing/v1/events/active
X-API-Key: <shared-secret>
```

```json
{
  "api_version": "v1",
  "generated_at": "2026-08-01T12:00:00Z",
  "event": {
    "id": 12,
    "name": "秋季スポーツ大会",
    "year": 2026,
    "season": "autumn",
    "status": "active",
    "start_date": "2026-09-01T00:00:00+09:00",
    "end_date": "2026-09-02T00:00:00+09:00"
  },
  "sports": [
    { "id": 8, "name": "タイピング" }
  ]
}
```

### 競技スナップショット

```http
GET /api/integrations/typing/v1/events/{event_id}/sports/{sport_id}
X-API-Key: <shared-secret>
```

次の情報を一度に返します。

- 大会ID、名称、年度、季節、状態、開催期間
- 競技ID、名称
- トーナメントID、名称
- 試合ID、ラウンド番号、試合番号、状態、開始・終了時刻、出場チームID
- チームID、名称、クラスID、クラス名
- 本登録済み選手のID、表示名、出場順

メールアドレス、ロール、ログイン情報は返しません。`entry_order` は1始まりで、タイピング競技の第1〜第3試合への出場順として使用します。

レスポンスの `warnings` には、RFPの前提（6チーム、各チーム3名）を満たさない場合や試合未登録の場合の診断情報が入ります。警告がある場合、タイピングシステムは運営画面へ表示し、試合開始前に確認できるようにしてください。

試合開始後にSportEase側の登録が変更されても競技中の割り当てが変わらないよう、タイピングシステムは開始操作時点のレスポンスをスナップショットとして保存してください。

### 出場順の設定（SportEase管理者用）

この操作は外部APIキーではなく、SportEaseの `admin` または `root` セッションと通常のCSRF保護を使用します。

```http
PUT /api/admin/typing/teams/{team_id}/entry-order
Content-Type: application/json

{
  "player_ids": ["player-id-for-match-1", "player-id-for-match-2", "player-id-for-match-3"]
}
```

`player_ids` は、そのチームで本登録済みの3選手と完全に一致する必要があります。配列順がそのまま `entry_order` 1、2、3になります。

## エラー

| HTTP状態 | 意味 |
| --- | --- |
| `400` | 大会ID・競技ID・チームID・リクエスト形式が不正 |
| `401` | APIキーがない、または一致しない |
| `404` | 操作対象大会、または指定された大会競技が存在しない |
| `409` | 出場順に指定した選手と本登録済み選手が一致しない |
| `429` | レート制限超過 |
| `503` | SportEaseに `TYPING_API_KEY` が未設定 |

## スコア連携の責務

このAPIは「SportEase → タイピングシステム」の大会・出場情報を提供します。逆方向の確定スコアについては、RFPどおりタイピングシステム側が取得APIを提供し、SportEaseが取得する構成です。そのAPIでは、再取得しても重複しない結果ID、スコア状態、無効・失格・再試合状態を含めてください。タイピングシステム側のURLと契約が確定後、SportEaseのスコア取り込み処理を接続します。
