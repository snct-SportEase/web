<script>
  import { browser } from '$app/environment';
  import { ensurePushSubscription, userHasPushEligibleRole } from '$lib/utils/push.js';
  import { onMount } from 'svelte';

  export let user;

  let isSubscribed = false;
  let isLoading = false;
  let permissionStatus = 'default';
  let isSupported = false;
  let errorMessage = '';
  let subscriptionCount = 0;

  $: canEnableNotifications = userHasPushEligibleRole(user);

  onMount(() => {
    if (browser) {
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
      const result = await ensurePushSubscription();
      
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
          errorMessage = '通知設定が不完全です。管理者にお問い合わせください。';
        } else {
          errorMessage = '通知の有効化に失敗しました。';
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

  // デバッグ情報を表示（開発環境のみ）
  function showDebugInfo() {
    if (!browser) return '';
    const debug = {
      isHTTPS: window.location.protocol === 'https:',
      origin: window.location.origin,
      notificationSupport: 'Notification' in window,
      serviceWorkerSupport: 'serviceWorker' in navigator,
      pushManagerSupport: 'PushManager' in window,
      permission: Notification.permission,
      isSupported: isSupported,
      isSubscribed: isSubscribed,
      subscriptionCount: subscriptionCount
    };
    console.log('[Notification Debug]', debug);
    return JSON.stringify(debug, null, 2);
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
    <div class="text-right">
      <span class="text-sm font-medium {getStatusColor()}">
        {getStatusText()}
      </span>
    </div>
  </div>

  {#if !canEnableNotifications}
    <div class="rounded-md border border-yellow-200 bg-yellow-50 px-4 py-3 text-sm text-yellow-800">
      通知機能は学生、管理者、ルートユーザーのみ利用できます。
    </div>
  {:else if !isSupported}
    <div class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
      このブラウザはプッシュ通知をサポートしていません。モバイルブラウザまたは最新のデスクトップブラウザをご利用ください。
    </div>
  {:else if errorMessage}
    <div class="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
      {errorMessage}
    </div>
  {/if}

  {#if canEnableNotifications && isSupported}
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
            on:click={disableNotifications}
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
            on:click={enableNotifications}
            disabled={isLoading || permissionStatus === 'denied'}
            class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {#if isLoading}
              処理中...
            {:else if permissionStatus === 'denied'}
              通知が拒否されています
            {:else}
              通知を有効にする
            {/if}
          </button>
        {/if}
      </div>
    </div>
  {/if}

  {#if browser && import.meta.env.DEV}
    <details class="mt-4">
      <summary class="cursor-pointer text-sm text-gray-500 hover:text-gray-700">
        デバッグ情報を表示
      </summary>
      <pre class="mt-2 rounded bg-gray-100 p-3 text-xs overflow-auto">{showDebugInfo()}</pre>
    </details>
  {/if}
</div>

