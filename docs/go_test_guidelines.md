# Goテスト規約

## 目的

SportEase の Go バックエンドにおけるテスト品質を安定させ、処理漏れ・例外漏れ・副作用漏れを防ぐための規約です。

この規約は特に以下を重視します。

- 1公開メソッドごとに責務が追えること
- 正常系だけでなく異常系・境界値が明示的に検証されること
- mock や sqlmock の期待値が厳密であること
- バグ修正時に再発防止テストが必ず追加されること

## 基本方針

- 1公開メソッドにつき、最低1つの親テスト関数を作る
- 分岐ごとに `t.Run(...)` でケースを分ける
- 正常系1件だけで終わらせない
- 副作用を持つ処理は、副作用の発生有無まで必ず検証する
- `mock.AssertExpectations(t)` または `mock.ExpectationsWereMet()` を必ず呼ぶ

推奨形式:

```go
func TestUserRepository_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {})
	t.Run("validation error", func(t *testing.T) {})
	t.Run("duplicate key", func(t *testing.T) {})
	t.Run("transaction rollback on insert failure", func(t *testing.T) {})
}
```

## 対象ごとの粒度

### Repository テスト

- 1メソッド1親テスト関数
- SQL の引数と実行順を確認する
- `sql.ErrNoRows` の扱いを確認する
- transaction を使う場合は `BEGIN` / `COMMIT` / `ROLLBACK` を確認する
- 更新系は「途中で失敗したら後続処理を実行しない」ことを確認する

例:

- `Create`
- `GetByID`
- `ListByEventID`
- `Update`
- `Delete`

### Handler テスト

- 1ハンドラメソッド1親テスト関数
- `status code`
- `response body`
- `repository 呼び出し`
- `repository が呼ばれないべきケース`
- 認証・認可失敗
- bind エラー
- 下位層エラー

を明示的に検証する

### Middleware テスト

- 認証あり
- 認証なし
- 権限あり
- 権限なし
- 不正トークン
- コンテキスト設定内容

を最低限確認する

## 命名規則

### テスト関数名

- `Test<Type>_<Method>`
- 例:
  - `TestMICRepository_VoteMIC`
  - `TestClassHandler_GetClassProgress`
  - `TestAuthMiddleware`

### サブテスト名

- 英語または日本語どちらでもよいが、プロジェクト内で一貫させる
- 状況と期待結果が分かる名前にする
- 抽象的な `case1`, `patternA` は禁止

推奨:

- `success`
- `not found`
- `invalid request body`
- `unauthorized`
- `repository returns error`
- `rollback when score log insert fails`

## 必須ケース

各公開メソッドは、内容に応じて最低でも以下を検討すること。

### 読み取り系

- 正常系
- データなし
- 下位層エラー
- 境界値

### 更新系・作成系

- 正常系
- 入力不正
- 対象なし
- 重複や制約違反
- 下位層エラー
- transaction rollback
- 後続副作用が発生しないこと

### 認証・認可が絡む処理

- 認証済み
- 未認証
- 権限あり
- 権限なし

## 期待値の厳格さ

### Repository

- `assert.NoError` / `assert.Error` だけで終わらせない
- 返却値の中身まで確認する
- `sqlmock` では SQL 実行順と引数を確認する
- 追加・更新・削除時は `RowsAffected` 相当の意味を確認する

### Handler

- body 全体、または主要フィールドを確認する
- JSON の shape が意図通りか確認する
- repository に渡した引数を mock で検証する

## Mock / sqlmock 運用ルール

### `testify/mock`

- 期待した呼び出しは `On(...).Return(...).Once()` を基本にする
- 呼ばれないべき処理は `AssertNotCalled` を使う
- 最後に `AssertExpectations(t)` を必ず呼ぶ

### `sqlmock`

- transaction を使う場合:
  - `ExpectBegin`
  - `ExpectQuery` / `ExpectExec`
  - `ExpectCommit` または `ExpectRollback`
- `mock.ExpectationsWereMet()` を必ず呼ぶ

## 例外系の扱い

- error が返るだけでなく、期待するメッセージまたは種類を確認する
- `nil, nil` を返す設計なら、その契約をテスト名に明記する
- panic を起こしてはいけないメソッドは、通常の error として処理されることを確認する

## テストデータの方針

- 1ケースに必要な最小限のデータだけ置く
- 使い回しが複雑になる helper は乱用しない
- ただし同一パターンが3回以上出る場合は helper 化を検討する

## してはいけないこと

- 正常系しか書かない
- `assert.Error(t, err)` だけで内容を見ない
- mock の期待値確認を省略する
- 1テストで複数公開メソッドをまとめて検証する
- 巨大なテストで複数責務を混ぜる
- 失敗時に原因が分からないサブテスト名にする

## 推奨ディレクトリ構成

- repository テスト: `backapp/tests/repository/*_test.go`
- handler テスト: `backapp/tests/handler/*_test.go`
- mock の共通定義: `backapp/tests/handler/main_test.go`

新しい責務が増えて mock が大きくなりすぎた場合は、責務別に mock ファイルを分割してよい。

## バグ修正時のルール

- バグを直す前に、まず再現テストを追加する
- 再現テストが落ちることを確認する
- 修正後にそのテストが通ることを確認する
- 可能なら近い境界ケースも追加する

## CI ルール

- `go test ./...` が通ることを必須とする
- backend CI で `go test -race ./...` を回す
- coverage を生成し、将来的にしきい値管理できる状態を維持する

## レビュー観点

Go のテストレビューでは、次を必ず確認する。

- この公開メソッドの主要分岐は全部カバーされているか
- 正常系以外の失敗パスがあるか
- 副作用は検証されているか
- mock が甘すぎないか
- ケース名だけで意図が読めるか
- 将来のリファクタで壊れたとき、どこが壊れたかすぐ分かるか

## このリポジトリでの最初の改善対象

既存コードでは、以下の順で厳格化するのを推奨する。

1. `repository` 層の更新系メソッド
2. `handler` 層の認証・認可つき API
3. 集計系・進行状況系の handler
4. middleware

特に transaction を使う repository は、rollback 漏れや後続副作用漏れが起きやすいため優先度が高い。
