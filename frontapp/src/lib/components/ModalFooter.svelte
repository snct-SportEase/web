<script>
  import Button from '$lib/components/Button.svelte';

  let {
    confirmLabel = '保存',
    cancelLabel = 'キャンセル',
    confirmDisabled = false,
    confirmLoading = false,
    onconfirm,
    oncancel
  } = $props();
</script>

<Button
  class="w-full sm:ml-3 sm:w-auto"
  onclick={onconfirm}
  disabled={confirmDisabled}
  loading={confirmLoading}
>
  {confirmLabel}
</Button>
{#if cancelLabel}
  <Button class="mt-3 w-full sm:mt-0 sm:w-auto" variant="secondary" onclick={oncancel}>
    {cancelLabel}
  </Button>
{/if}
