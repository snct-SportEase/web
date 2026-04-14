# フロントエンドテスト規約

## 目的

SportEase のフロントエンドにおける品質を安定させ、UIのデグレード、ロジックのバグ、およびクロスブラウザでの動作不良を防ぐための規約です。

この規約は特に以下を重視します。

- 複雑なビジネスロジックが独立して検証されていること
- UIコンポーネントが期待通りにレンダリングされ、ユーザー操作に反応すること
- 実際のブラウザ環境で重要なユーザーフローが正常に動作すること (E2E)
- バグ修正時に再発防止テストが必ず追加されること

## 基本方針

- **ロジックテスト**: 純粋なJavaScript/TypeScript関数は、ブラウザ環境に依存しないNode.js環境で高速に実行する。
- **コンポーネントテスト**: Svelteコンポーネントは `vitest` のブラウザモードを利用し、実際のブラウザ（Chromium等）でレンダリングとインタラクションを検証する。
- **E2Eテスト**: 複数の画面をまたぐ操作や、実際のバックエンドAPI（またはモック）との連携を含む重要なフローは `Playwright` で検証する。
- **アクセシビリティ**: 要素の取得には可能な限り `getByRole`, `getByLabelText` 等のアクセシビリティに基づいたクエリを使用する。

## テストの種類とツール

| 種類 | ツール | 実行環境 | 対象 |
| :--- | :--- | :--- | :--- |
| ロジックテスト | Vitest | Node.js | ヘルパー関数、ストア、計算ロジック |
| コンポーネントテスト | Vitest (Browser) | Browser | Svelteコンポーネント単体、UIインタラクション |
| E2Eテスト | Playwright | Browser | ログイン、エントリー登録、大会進行等の重要フロー |

## 命名規則と配置

### ファイル名

- ロジック/コンポーネントテスト: `*.spec.js` または `*.test.js`
- コンポーネントテスト（Svelte）: `*.svelte.spec.js`
- E2Eテスト: `e2e/*.spec.js`

### 配置

- **ロジック/コンポーネント**: テスト対象のファイルと同じディレクトリに配置する（Co-location）。
  - 例: `src/lib/utils/date.js` -> `src/lib/utils/date.spec.js`
  - 例: `src/lib/components/Button.svelte` -> `src/lib/components/Button.svelte.spec.js`
- **E2E**: `frontapp/e2e/` ディレクトリにまとめて配置する。

## Vitest ブラウザモード (コンポーネントテスト)

Svelteコンポーネントのテストには `vitest-browser-svelte` を使用します。

### 基本的な書き方

```javascript
import { page } from '@vitest/browser/context';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import MyComponent from './MyComponent.svelte';

describe('MyComponent', () => {
  it('初期表示が正しいこと', async () => {
    render(MyComponent, { props: { title: 'Hello' } });
    
    const heading = page.getByRole('heading', { name: 'Hello' });
    await expect.element(heading).toBeInTheDocument();
  });

  it('ボタンクリックでイベントが発火すること', async () => {
    const { component } = render(MyComponent);
    // モック関数の作成などは Vitest の標準機能を使用
    
    const button = page.getByRole('button', { name: '送信' });
    await button.click();
    
    // 期待される結果の検証
  });
});
```

## Playwright (E2Eテスト)

システムの主要なユースケースを検証します。

### 基本的な書き方

```javascript
import { test, expect } from '@playwright/test';

test('ユーザーがログインしてトップページを表示できる', async ({ page }) => {
  await page.goto('/');
  
  // ログイン操作のシミュレーション
  await page.getByLabel('ユーザー名').fill('test-user');
  await page.getByLabel('パスワード').fill('password');
  await page.getByRole('button', { name: 'ログイン' }).click();
  
  // 遷移後の確認
  await expect(page).toHaveURL('/dashboard');
  await expect(page.getByRole('heading', { name: 'ダッシュボード' })).toBeVisible();
});
```

## テストの実行方法

`frontapp` ディレクトリで以下のコマンドを実行します。

- **ユニット/コンポーネントテスト（全件）**: `npm run test:unit`
- **E2Eテスト**: `npm run test:e2e`
- **CI用（エラーフィルタリングあり）**: `npm run test`
- **特定のファイルを実行**: `npx vitest src/path/to/file.spec.js`

## 必須ケース

### UIコンポーネント
- 正しくレンダリングされること（デフォルト状態）
- Propsの変化に応じて表示が変わること
- ユーザー操作（クリック、入力）に対して期待通りのイベントや状態変化が起きること
- ローディング中、エラー時の表示

### ビジネスロジック
- 正常系
- 異常系（不正な入力、APIエラー等）
- 境界値（空配列、極端に長い文字列、0、負の数等）

### E2E
- ログイン/ログアウト
- 主要なデータの登録・編集・削除フロー
- 権限による画面アクセスの制限

## してはいけないこと

- **実装詳細のテスト**: `component.instance().someMethod()` のような、内部状態やプライベートメソッドを直接呼ぶテストは避ける。ユーザーが見るもの、操作するものを基準にする。
- **過度なモック**: 可能な限り実際の挙動に近い形でテストする。ただし、バックエンドAPIなどは必要に応じてモックする。
- **スリープの多用**: `setTimeout` 等で待つのではなく、`expect.element(...).toBeInTheDocument()` や Playwright の `waitFor` 系メソッドを使用して、状態の変化を待つ。
- **テスト間の依存**: 各テストケースは独立して実行可能であること。

## レビュー観点

- `getByRole` 等のセマンティックなクエリが使われているか
- 異常系や境界値の考慮が漏れていないか
- テスト名が「何を確認しているか」明確か
- 非同期処理（API呼び出しやレンダリング待ち）が適切に `await` されているか
- 再現したバグに対するテストコードが含まれているか（修正時）

## CI ルール

- PR作成時に GitHub Actions で `npm run test` (Vitest) および `npm run test:e2e` が実行される。
- すべてのテストがパスすることがマージの条件となる。
