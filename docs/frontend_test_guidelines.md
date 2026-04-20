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

## 機能別テスト進捗表

テストの実施状況を以下の表で管理します。
- ○: テスト実装済み
- ☓: 未実装

### root（最上位管理者）機能

| 機能 | 概要 | コンポーネント | E2E | 備考 |
| :--- | :--- | :---: | :---: | :--- |
| 大会の作成・編集 | 年度・シーズン・期間等の設定 | ○ | ○ | `event-management.svelte.spec.js` / `root-event-management.spec.js` で確認済み |
| 大会ステータス変更 | upcoming / active / archived の切り替え | ○ | ○ | 編集保存時の `status` 更新を確認済み |
| スコア非表示設定 | 学生へのスコア公開・非公開設定 | ○ | ○ | `event-management.svelte.spec.js` / `root-event-management.spec.js` で確認済み |
| アンケート通知配信 | アンケートURLの全体通知送信 | ○ | ○ | `event-management.svelte.spec.js` / `root-event-management.spec.js` で通知送信フローを確認済み |
| 得点CSVインポート | 外部集計データのインポート（春季） | ○ | ○ | `event-management.svelte.spec.js` / `root-event-management.spec.js` で確認済み |
| 結果のCSV/PDF出力 | クラス別スコア集計の出力 | ○ | ○ | `event-management.svelte.spec.js` / `root-event-management.spec.js` で確認済み |
| DBダンプ出力 | データベース全体のエクスポート | ○ | ○ | `event-management.svelte.spec.js` / `root-event-management.spec.js` で確認済み |
| 雨天時モード切替 | 競技中止・敗者復活戦追加の一括制御 | ○ | ○ | `rainy-mode.svelte.spec.js` / `root-rainy-mode.spec.js` で確認済み |
| 雨天時定員設定 | 競技・クラスごとの雨天時定員設定 | ○ | ○ | `sport-details-registration.svelte.spec.js` / `root-rainy-capacity-settings.spec.js` で確認済み |
| 競技マスタ登録 | システム共通の競技種目登録 | ○ | ○ | `sport-management.svelte.spec.js` / `root-sport-management.spec.js` で確認済み |
| 大会への競技割り当て | 競技の大会紐付け・ルール設定 | ○ | ○ | `sport-management.svelte.spec.js` / `root-sport-management.spec.js` で確認済み |
| 通知作成・配信 | プッシュ通知の作成と送信 | ○ | ○ | `notification.svelte.spec.js` / `root-notification.spec.js` で確認済み |
| 通知申請の承認・否認 | 学生からの通知申請の審査 | ○ | ○ | `notification-requests.svelte.spec.js` / `root-notification-requests.spec.js` で確認済み |
| 権限管理 | `admin` / `root` 権限の付与・剥奪 | ☓ | ☓ | `user-promotion` 画面の自動テストは未整備 |
| トーナメント自動生成 | ブラケットの自動生成と確定 | ○ | ○ | `tournament-management.svelte.spec.js` / `root-tournament-management.spec.js` で確認済み |
| 昼競技セッション管理 | 昼競技の開催枠・ポイント設定 | ○ | ○ | `noon-game.svelte.spec.js` / `root-noon-game.spec.js` で確認済み |
| 昼競技テンプレート実行 | リレー等の対戦カード自動生成 | ○ | ○ | `noon-game.svelte.spec.js` / `root-noon-game.spec.js` で確認済み |
| ユーザー表示名変更 | ユーザーの表示名更新 | ○ | ○ | `change-username.svelte.spec.js` / `root-change-username.spec.js` で確認済み |
| クラス所属ロール付け替え | クラス所属（クラス名_rep）の変更 | ○ | ○ | `change-username.svelte.spec.js` / `root-change-username.spec.js` で確認済み |
| クラス人数設定 | 各クラスの学生数更新 | ○ | ○ | `class-student-count.svelte.spec.js` / `root-class-student-count.spec.js` で確認済み |
| MIC結果確認 | 投票結果の確認 | ○ | ○ | `identify-mic.svelte.spec.js` / `root-identify-mic.spec.js` で確認済み |
| 競技要項PDFアップロード | 競技要項PDFの登録・管理 | ○ | ○ | `competition-guidelines-upload.svelte.spec.js` / `root-competition-guidelines-upload.spec.js` で確認済み |

### admin（運営スタッフ）機能

| 機能 | 概要 | コンポーネント | E2E | 備考 |
| :--- | :--- | :---: | :---: | :--- |
| 管理者ダッシュボード閲覧 | 統計情報のリアルタイム確認 | ☓ | ☓ | |
| チームメンバー割り当て | 競技参加メンバーの登録・削除 | ☓ | ☓ | |
| ロール付与・削除 | 審判ロール等の付与 | ☓ | ☓ | |
| QRコードスキャン | 学生QRのスキャン・参加確認 | ☓ | ☓ | |
| 参加確認済み一覧 | スキャン済み学生の確認 | ☓ | ☓ | |
| 出席者数の登録 | クラスごとの出席数入力 | ☓ | ☓ | |
| 競技詳細の登録・更新 | ルール・定員・開始時間の設定 | ☓ | ☓ | |
| 試合ステータスの更新 | 試合の進行状況変更 | ☓ | ☓ | |
| 試合結果の入力 | トーナメントスコア・勝敗入力 | ☓ | ☓ | |
| 昼競技結果の入力 | 昼競技の結果登録 | ☓ | ☓ | |
| MIC投票 | MIC候補への投票 | ☓ | ☓ | |

### student（学生）機能

| 機能 | 概要 | コンポーネント | E2E | 備考 |
| :--- | :--- | :---: | :---: | :--- |
| マイページの閲覧 | 自クラスのスコア・日程確認 | ☓ | ☓ | |
| トーナメントの閲覧 | ブラケットと結果の確認 | ☓ | ☓ | |
| タイムテーブルの閲覧 | 試合スケジュールの確認 | ☓ | ☓ | |
| 昼競技情報の閲覧 | 昼競技の結果・ポイント確認 | ☓ | ☓ | |
| 競技情報の閲覧 | ルール・開催場所の確認 | ☓ | ☓ | |
| スコア一覧の閲覧 | 全クラスのランキング確認 | ☓ | ☓ | |
| QRコードの発行 | 参加確認用QRの生成 | ☓ | ☓ | |
| 通知の受信・フィルタ設定 | プッシュ通知の受信と表示管理 | ☓ | ☓ | |
| プッシュ通知の設定 | ブラウザ通知の有効化・解除 | ☓ | ☓ | |
| 通知申請の提出 | 運営への通知配信申請 | ☓ | ☓ | |
| 申請状況の確認 | 提出済み申請のステータス確認 | ☓ | ☓ | |
| 過去大会の閲覧 | アーカイブデータの参照 | ☓ | ☓ | |
| ガイドの閲覧 | PWAインストール・要項確認 | ☓ | ☓ | |
| クラス情報の閲覧 | 出席状況・メンバー一覧の確認 | ☓ | ☓ | |
