<script>
  import '../app.css';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import { page } from '$app/stores';
  import { isSidebarOpen } from '$lib/stores/sidebarStore.js';
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  $: data = $page.data;

  onMount(() => {
    if (browser && 'serviceWorker' in navigator) {
      window.addEventListener('load', () => {
        navigator.serviceWorker.register('/service-worker.js');
      });
    }
  });

  function openSidebar() {
    isSidebarOpen.set(true);
  }
</script>

<div class="app-container">
  {#if data.user}
    <Sidebar user={data.user} />
  {/if}
  <main class="main-content">
    {#if !$isSidebarOpen}
      <button on:click={openSidebar} class="open-sidebar-button">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path></svg>
      </button>
    {/if}
    <slot />
  </main>
</div>
