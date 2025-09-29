<script>
  import ProfileSetupModal from '$lib/components/ProfileSetupModal.svelte';

  let { data, children } = $props();

</script>

{#if data.userProfile?.is_profile_complete}
  <!-- プロフィールが完成している場合は、子ページ（/dashboard/+page.svelteなど）を表示 -->
  {@render children()}
{:else}
  <!-- プロフィールが未完成の場合は、設定モーダルを表示 -->
  <ProfileSetupModal classes={data.classes} form={data.form} />
{/if}
