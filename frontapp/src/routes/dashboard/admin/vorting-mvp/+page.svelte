<script>
  import { onMount } from 'svelte';

  let eligibleClasses = [];
  let selectedClass = '';
  let reason = '';
  let eventId = 1; // TODO: Get the active event id

  let hasVoted = false;
  let votedForClassId = null;
  let votedForClassName = '';

  onMount(async () => {
    // Fetch eligible classes first to map class ID to name
    const classRes = await fetch(`/api/admin/mvp/eligible-classes?event_id=${eventId}`);
    if (classRes.ok) {
      eligibleClasses = await classRes.json();
    } else {
      console.error('Failed to fetch eligible classes');
      // Handle error appropriately
    }

    // Check if user has already voted
    const voteRes = await fetch(`/api/admin/mvp/user-vote?event_id=${eventId}`);
    if (voteRes.ok) {
      const voteData = await voteRes.json();
      if (voteData.voted) {
        hasVoted = true;
        votedForClassId = voteData.vote.voted_for_class_id;
        const votedClass = eligibleClasses.find(c => c.id === votedForClassId);
        if (votedClass) {
          votedForClassName = votedClass.name;
        }
      }
    } else {
      console.error('Failed to fetch user vote status');
    }
  });

  async function vote() {
    if (!selectedClass) {
      alert('Please select a class to vote for.');
      return;
    }

    const res = await fetch('/api/admin/mvp/vote', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        voted_for_class_id: parseInt(selectedClass),
        reason: reason,
        event_id: eventId,
      }),
    });

    if (res.ok) {
      alert('Vote successful!');
      hasVoted = true;
      votedForClassId = parseInt(selectedClass);
      const votedClass = eligibleClasses.find(c => c.id === votedForClassId);
      if (votedClass) {
        votedForClassName = votedClass.name;
      }
    } else {
      try {
        const error = await res.json();
        alert(`Error: ${error.error}`);
      } catch (e) {
        alert('An unknown error occurred.');
      }
    }
  }
</script>

<h1 class="text-2xl font-bold mb-4">MVP投票</h1>

{#if hasVoted}
  <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6 text-center">
    <h2 class="text-xl font-bold mb-2">投票済みです</h2>
    <p>あなたは <span class="font-bold">{votedForClassName}</span> に投票しました。</p>
    <p class="text-gray-600 mt-4">MVP投票は一人一票までです。</p>
  </div>
{:else}
  <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6">
    <!-- The form from before -->
    <div class="mb-4">
      <label for="class-select" class="block text-gray-700 font-bold mb-2">Eligible Class:</label>
      <select id="class-select" bind:value={selectedClass} class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline">
        <option value="" disabled>Select a class</option>
        {#each eligibleClasses as c}
          <option value={c.id}>{c.name}</option>
        {/each}
      </select>
    </div>

    <div class="mb-6">
      <label for="reason" class="block text-gray-700 font-bold mb-2">Reason:</label>
      <textarea id="reason" bind:value={reason} class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" rows="4" placeholder="Enter your reason here..."></textarea>
    </div>

    <div class="flex items-center justify-between">
      <button on:click={vote} class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline" type="button">
        Vote
      </button>
    </div>
  </div>
{/if}
