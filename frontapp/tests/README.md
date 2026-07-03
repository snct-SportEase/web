# Frontend Tests

フロントエンドのテストはこのディレクトリにまとめます。

- `unit/`: Vitest のロジック・Svelteコンポーネント・ページ単体テスト
- `e2e/`: Playwright のE2Eテスト

`unit/` 配下では `src/` からの相対パスを保ち、テスト対象の import には `$src/...` エイリアスを使います。

