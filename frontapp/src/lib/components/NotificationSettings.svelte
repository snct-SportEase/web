<script>
  import { browser } from '$app/environment';
  import { ensurePushSubscription, userHasPushEligibleRole } from '$lib/utils/push.js';
  import { isPWAInstalled, getDeviceType } from '$lib/utils/pwa.js';
  import { env as publicEnv } from '$env/dynamic/public';
  import { onMount } from 'svelte';

  let { user } = $props();

  let isSubscribed = $state(false);
  let isLoading = $state(false);
  let permissionStatus = $state('default');
  let isSupported = $state(false);
  let errorMessage = $state('');
  let subscriptionCount = $state(0);
  let isIOS = $state(false);
  let isPWA = $state(false);

  let canEnableNotifications = $derived(userHasPushEligibleRole(user));
  let vapidKeySet = $derived(browser ? (publicEnv.PUBLIC_WEBPUSH_PUBLIC_KEY ?? publicEnv.PUBLIC_WEBPUSH_KEY ?? '') !== '' : false);

  onMount(() => {
    if (browser) {
      isIOS = getDeviceType() === 'ios';
      isPWA = isPWAInstalled();
      checkNotificationSupport();
      checkPermissionStatus();
      loadSubscriptionStatus();
    }
  });

  function checkNotificationSupport() {
    isSupported = 'Notification' in window && 
                  'serviceWorker' in navigator && 
                  'PushManager' in window;
  }

  function checkPermissionStatus() {
    if (browser && 'Notification' in window) {
      permissionStatus = Notification.permission;
    }
  }

  async function loadSubscriptionStatus() {
    if (!browser || !canEnableNotifications) {
      return;
    }

    try {
      const response = await fetch('/api/notifications/subscription', {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        isSubscribed = data.subscribed || false;
        subscriptionCount = data.count || 0;
      }
    } catch (error) {
      console.error('[notification] Failed to load subscription status:', error);
    }
  }

  async function enableNotifications() {
    if (!browser || !isSupported) {
      errorMessage = 'このブラウザは通知をサポートしていません。';
      return;
    }

    if (permissionStatus === 'denied') {
      errorMessage = '通知が拒否されています。ブラウザの設定から通知を許可してください。';
      return;
    }

    isLoading = true;
    errorMessage = '';

    try {
      console.log('[notification] 通知の有効化を開始します');
      const result = await ensurePushSubscription();
      console.log('[notification] 通知の有効化結果:', result);
      
      if (result.status === 'subscribed') {
        isSubscribed = true;
        await loadSubscriptionStatus();
        errorMessage = '';
      } else if (result.status === 'skipped') {
        if (result.reason === 'permission-denied') {
          errorMessage = '通知の許可が必要です。ブラウザの設定を確認してください。';
          permissionStatus = 'denied';
        } else if (result.reason === 'unsupported') {
          errorMessage = 'このブラウザは通知をサポートしていません。';
        } else if (result.reason === 'missing-vapid-key') {
          errorMessage = '通知機能は現在利用できません。管理者にお問い合わせください。';
          console.error('[notification] VAPID public key is not configured.');
        } else {
          errorMessage = `通知の有効化に失敗しました。理由: ${result.reason || '不明'}`;
        }
      } else {
        errorMessage = result.reason || '通知の有効化に失敗しました。';
      }
    } catch (error) {
      console.error('[notification] Failed to enable notifications:', error);
      errorMessage = '通知の有効化中にエラーが発生しました。';
    } finally {
      isLoading = false;
      checkPermissionStatus();
    }
  }

  async function disableNotifications() {
    if (!browser) {
      return;
    }

    isLoading = true;
    errorMessage = '';

    try {
      // 現在の購読を取得
      const response = await fetch('/api/notifications/subscription', {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        const endpoints = data.endpoints || [];

        // 各エンドポイントの購読を削除
        for (const endpoint of endpoints) {
          try {
            await fetch('/api/notifications/subscription', {
              method: 'DELETE',
              headers: {
                'Content-Type': 'application/json'
              },
              credentials: 'include',
              body: JSON.stringify({ endpoint })
            });
          } catch (error) {
            console.error('[notification] Failed to delete subscription:', error);
          }
        }

        // Service Workerからも購読を解除
        if ('serviceWorker' in navigator) {
          try {
            const registration = await navigator.serviceWorker.ready;
            const subscription = await registration.pushManager.getSubscription();
            if (subscription) {
              await subscription.unsubscribe();
            }
          } catch (error) {
            console.error('[notification] Failed to unsubscribe from service worker:', error);
          }
        }

        isSubscribed = false;
        subscriptionCount = 0;
        errorMessage = '';
      }
    } catch (error) {
      console.error('[notification] Failed to disable notifications:', error);
      errorMessage = '通知の無効化中にエラーが発生しました。';
    } finally {
      isLoading = false;
    }
  }

  function getStatusText() {
    if (!isSupported) {
      return 'サポートされていません';
    }
    if (permissionStatus === 'denied') {
      return '通知が拒否されています';
    }
    if (permissionStatus === 'default') {
      return '未設定';
    }
    if (isSubscribed) {
      return '有効';
    }
    return '無効';
  }

  function getStatusColor() {
    if (!isSupported || permissionStatus === 'denied') {
      return 'text-red-600';
    }
    if (isSubscribed) {
      return 'text-green-600';
    }
    return 'text-gray-600';
  }

</script>

<div class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
  <div class="flex items-center justify-between mb-4">
    <div>
      <h3 class="text-lg font-semibold text-gray-800">プッシュ通知設定</h3>
      <p class="mt-1 text-sm text-gray-600">
        モバイルデバイスで通知を受け取ることができます
      </p>
    </div>
    {#if isSupported}
    <div class="text-right">
      <span class="text-sm font-medium {getStatusColor()}">
        {getStatusText()}
      </span>
    </div>
    {/if}
  </div>

  {#if !canEnableNotifications}
    <div class="rounded-md border border-yellow-200 bg-yellow-50 px-4 py-3 text-sm text-yellow-800">
      通知機能は学生、管理者、ルートユーザーのみ利用できます。
    </div>
  {:else if !isSupported}
    {#if isIOS && !isPWA}
      <div class="rounded-md border border-blue-200 bg-blue-50 px-4 py-3 text-sm text-blue-800">
        <p class="font-semibold mb-1">iOSでプッシュ通知を受け取るには、ホーム画面への追加が必要です</p>
        <ol class="list-decimal list-inside space-y-1 mt-2">
          <li>SafariでこのページをSafariで開く</li>
          <li>画面下の共有ボタン（四角に矢印のアイコン）をタップ</li>
          <li>「ホーム画面に追加」を選択</li>
          <li>ホーム画面のアイコンからアプリを開き、通知を有効化する</li>
        </ol>
      </div>
    {:else}
      <div class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        このブラウザはプッシュ通知をサポートしていません。モバイルブラウザまたは最新のデスクトップブラウザをご利用ください。
      </div>
    {/if}
  {:else if errorMessage}
    <div class="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
      {errorMessage}
    </div>
  {/if}

  {#if canEnableNotifications && isSupported}
    {#if !vapidKeySet}
      <div class="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        <p class="font-semibold">通知機能は現在利用できません</p>
        <p class="mt-1">管理者にお問い合わせください。</p>
      </div>
    {/if}
    <div class="flex items-center justify-between">
      <div class="flex-1">
        {#if isSubscribed}
          <p class="text-sm text-gray-700">
            通知は有効です。{subscriptionCount > 0 ? `${subscriptionCount}件のデバイスで通知を受信できます。` : ''}
          </p>
        {:else}
          <p class="text-sm text-gray-700">
            通知を有効にすると、重要な情報をリアルタイムで受け取ることができます。
          </p>
        {/if}
      </div>
      <div class="ml-4">
        {#if isSubscribed}
          <button
            type="button"
            onclick={disableNotifications}
            disabled={isLoading}
            class="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {#if isLoading}
              処理中...
            {:else}
              通知を無効にする
            {/if}
          </button>
        {:else}
          <button
            type="button"
            onclick={enableNotifications}
            disabled={isLoading || permissionStatus === 'denied' || !vapidKeySet}
            class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {#if isLoading}
              処理中...
            {:else if permissionStatus === 'denied'}
              通知が拒否されています
            {:else if !vapidKeySet}
              利用できません
            {:else}
              通知を有効にする
            {/if}
          </button>
        {/if}
      </div>
    </div>
  {/if}

</div>
