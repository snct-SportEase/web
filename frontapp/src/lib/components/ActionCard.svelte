<script>
  let {
    title,
    description,
    actionLabel = '詳細を見る',
    href = null,
    onclick,
    size = 'compact',
    headingTag = 'h3',
    children,
    class: className = ''
  } = $props();

  const paddingClass = $derived(size === 'spacious' ? 'p-6' : 'p-5');
  const titleClass = $derived(size === 'spacious' ? 'text-xl' : 'text-base');
</script>

<svelte:element
  this={href ? 'a' : 'button'}
  href={href || undefined}
  type={href ? undefined : 'button'}
  role={href ? 'link' : 'button'}
  {onclick}
  class={`group block rounded-lg border border-indigo-100 bg-white text-left shadow-sm transition hover:border-indigo-300 hover:shadow ${paddingClass} ${className}`}
>
  {#if children}
    <div class={size === 'spacious' ? 'mb-3 flex items-center' : 'mb-2 flex items-center'}>
      {@render children()}
    </div>
  {/if}
  <svelte:element this={headingTag} class={`${titleClass} font-semibold text-indigo-700 group-hover:text-indigo-800`}>
    {title}
  </svelte:element>
  <p class="mt-1 text-sm text-gray-600">{description}</p>
  <span class="mt-3 inline-flex items-center text-sm font-medium text-indigo-600 group-hover:text-indigo-700">
    {actionLabel} →
  </span>
</svelte:element>
