# 通知が来ない問題のトラブルシューティングガイド

## 考えられる原因

### 1. **HTTPSの問題** ⚠️ 最重要
プッシュ通知は**HTTPS必須**です（localhostは例外）。
- 本番環境がHTTPの場合、通知は動作しません
- ブラウザの開発者ツールでコンソールエラーを確認
- `navigator.serviceWorker`が利用可能か確認

### 2. **VAPIDキーの設定不一致**
- **フロントエンド**: `PUBLIC_WEBPUSH_PUBLIC_KEY` が設定されているか
- **バックエンド**: `WEBPUSH_PUBLIC_KEY` と `WEBPUSH_PRIVATE_KEY` が設定されているか
- **重要**: フロントエンドとバックエンドで**同じVAPIDキーペア**を使用する必要があります
- 環境変数が正しく読み込まれているか確認

### 3. **Service Workerの登録問題**
- Service Workerが正しく登録されているか
- ブラウザの開発者ツール > Application > Service Workers で確認
- Service Workerのスコープが正しいか確認

### 4. **ドメイン/オリジンの不一致**
- VAPIDキーは特定のドメインに紐づいています
- 本番環境のドメインで生成したVAPIDキーを使用しているか確認
- ローカルと本番で異なるVAPIDキーを使用している可能性

### 5. **購読情報の保存/取得問題**
- 購読情報がデータベースに正しく保存されているか
- ユーザーIDとエンドポイントの紐付けが正しいか
- 複数のデバイスで購読している場合、すべてのエンドポイントが取得できているか

### 6. **バックエンドの送信エラー**
- バックエンドのログを確認
- `[notification] Push送信に失敗しました` のエラーログがないか
- VAPIDキーが正しく設定されているか（空文字列でないか）

### 7. **ブラウザの通知許可状態**
- ブラウザの通知が拒否されていないか
- ブラウザの設定で通知がブロックされていないか
- プライベートモードや厳格なプライバシー設定の影響

### 8. **ネットワーク/ファイアウォールの問題**
- プッシュ通知サービス（FCM等）への接続がブロックされていないか
- 企業ネットワークやプロキシの影響

## 診断手順

### ステップ0: バックエンドのログを確認（最重要）
通知送信時に以下のログが出力されます。**本番環境のバックエンドログを確認してください**：

```
[notification] 通知送信開始: notificationID=1, title=..., targetRoles=[...]
[notification] 対象ユーザー数: X, userIDs=[...]
[notification] 購読情報数: X
[notification] X件の購読に対してPush通知を送信します
[notification] [1/X] Push送信試行: userID=..., endpoint=...
[notification] [1/X] Push送信成功: userID=..., endpoint=..., status=201
```

**よくある問題**：
- `VAPIDキーが設定されていないためPush通知をスキップします` → 環境変数を確認
- `対象ユーザーが0人のためPush通知をスキップします` → ロールの設定を確認
- `購読情報が0件のためPush通知をスキップします` → ユーザーが通知を有効化しているか確認
- `Push送信に失敗しました` → エラーメッセージを確認

### ステップ1: ブラウザのコンソールで確認
```javascript
// ブラウザの開発者ツールのコンソールで実行
console.log('Notification support:', 'Notification' in window);
console.log('Service Worker support:', 'serviceWorker' in navigator);
console.log('Push Manager support:', 'PushManager' in window);
console.log('Notification permission:', Notification.permission);
console.log('Current origin:', window.location.origin);
console.log('Is HTTPS:', window.location.protocol === 'https:');

// Service Workerの状態確認
navigator.serviceWorker.getRegistrations().then(registrations => {
  console.log('Service Worker registrations:', registrations);
  registrations.forEach(reg => {
    reg.pushManager.getSubscription().then(sub => {
      console.log('Push subscription:', sub ? sub.toJSON() : 'None');
    });
  });
});
```

### ステップ2: フロントエンドの環境変数確認
- ブラウザのコンソールで `PUBLIC_WEBPUSH_PUBLIC_KEY` が設定されているか確認
- ネットワークタブで `/api/notifications/subscription` のリクエスト/レスポンスを確認

### ステップ3: バックエンドのログ確認
通知送信時に以下のログが出力されるか確認：
- `[notification] VAPIDキーが設定されていないためPush通知をスキップします`
- `[notification] ユーザー抽出に失敗しました`
- `[notification] 購読情報の取得に失敗しました`
- `[notification] Push送信に失敗しました`

### ステップ4: データベースの確認
```sql
-- 購読情報が保存されているか確認
SELECT * FROM push_subscriptions WHERE user_id = 'ユーザーID';

-- 通知が作成されているか確認
SELECT * FROM notifications ORDER BY created_at DESC LIMIT 10;
```

### ステップ5: ネットワークリクエストの確認
- ブラウザの開発者ツール > Network タブで確認
- `/api/notifications/subscription` のPOSTリクエストが成功しているか
- レスポンスのステータスコードが201か確認

### ステップ6: 診断エンドポイントで確認
以下のエンドポイントで詳細な診断情報を取得できます：

```bash
# ブラウザのコンソールで実行
fetch('/api/notifications/debug', { credentials: 'include' })
  .then(r => r.json())
  .then(console.log);
```

レスポンス例：
```json
{
  "user_id": "user-123",
  "subscription_count": 2,
  "vapid_key_configured": true,
  "vapid_public_key_set": true,
  "subscriptions": [
    {
      "id": 1,
      "endpoint": "https://fcm.googleapis.com/...",
      "created_at": "2025-12-06T..."
    }
  ]
}
```

**確認ポイント**：
- `subscription_count` が0より大きいか（通知を有効化しているか）
- `vapid_key_configured` が `true` か（VAPIDキーが設定されているか）
- `subscriptions` にエンドポイントが含まれているか

## よくある問題と解決方法

### 問題1: VAPIDキーが設定されていない
**症状**: ブラウザコンソールに `missing-vapid-key` エラー
**解決方法**: 
- フロントエンドの `.env` に `PUBLIC_WEBPUSH_PUBLIC_KEY` を設定
- バックエンドの `.env` に `WEBPUSH_PUBLIC_KEY` と `WEBPUSH_PRIVATE_KEY` を設定
- 本番環境の環境変数も同様に設定

### 問題2: HTTPSでない
**症状**: Service Workerが登録されない、またはプッシュ通知が動作しない
**解決方法**: 
- 本番環境をHTTPSで提供する（Let's Encrypt等を使用）
- 開発環境ではlocalhostを使用（HTTPでも動作）

### 問題3: 異なるVAPIDキーを使用している
**症状**: 購読は成功するが通知が届かない
**解決方法**: 
- フロントエンドとバックエンドで同じVAPIDキーペアを使用
- 本番環境用のVAPIDキーを生成し、両方に設定

### 問題4: 購読情報が保存されていない
**症状**: 通知設定で「有効」と表示されるが通知が届かない
**解決方法**: 
- データベースに購読情報が保存されているか確認
- `/api/notifications/subscription` のPOSTリクエストが成功しているか確認
- バックエンドのログでエラーがないか確認
- 診断エンドポイント `/api/notifications/debug` で購読情報を確認

### 問題6: 対象ユーザーが0人
**症状**: バックエンドログに「対象ユーザーが0人のためPush通知をスキップします」
**解決方法**: 
- 通知の送信先ロールが正しく設定されているか確認
- ユーザーに該当するロールが割り当てられているか確認
- ロール名の大文字小文字が一致しているか確認

### 問題7: 購読情報が0件
**症状**: バックエンドログに「購読情報が0件のためPush通知をスキップします」
**解決方法**: 
- ユーザーが通知を有効化しているか確認
- 通知設定ページで「通知を有効にする」ボタンをクリックしているか確認
- データベースの `push_subscriptions` テーブルにレコードがあるか確認

### 問題5: 通知許可が拒否されている
**症状**: ブラウザの通知許可が「拒否」になっている
**解決方法**: 
- ブラウザの設定から通知を許可
- サイトの設定で通知を許可
- プライベートモードを解除

## VAPIDキーの生成方法

```bash
# OpenSSLを使用してVAPIDキーを生成
openssl ecparam -genkey -name prime256v1 -noout -out vapid_private_key.pem
openssl ec -in vapid_private_key.pem -pubout -out vapid_public_key.pem

# Base64 URL Safe形式に変換（Node.jsスクリプト等を使用）
# または、web-pushライブラリを使用して生成
```

## デバッグ用コードの追加

通知設定コンポーネントにデバッグ情報を表示する機能を追加することを推奨します。

