<script>
  import { onMount } from 'svelte';

  export let data;
  let { classes, managedClass } = data;

  let selectedClassId = null;
  let classDetails = null;

  let attendanceCount = 0;
  let isLoading = false;
  let errorMessage = '';
  let successMessage = '';

  onMount(() => {
    if (managedClass) {
      selectedClassId = managedClass.id;
      fetchClassDetails(managedClass.id);
    } else if (classes && classes.length > 0) {
      selectedClassId = classes[0].id;
      fetchClassDetails(selectedClassId);
    }
  });

  async function fetchClassDetails(classId) {
    if (!classId) return;

    isLoading = true;
    try {
      // The backend endpoint for class details seems to be missing from the provided file structure.
      // Assuming an endpoint exists at `/api/admin/attendance/class-details/${classId}` based on the original code.
      const response = await fetch(`/api/admin/attendance/class-details/${classId}`);
      if (response.ok) {
        classDetails = await response.json();
      } else {
        const errData = await response.json();
        errorMessage = errData.error || 'クラス詳細の取得に失敗しました。';
        classDetails = null; // Clear details on failure
      }
    } catch (err) {
      errorMessage = '通信エラーが発生しました。';
      classDetails = null; // Clear details on error
    } finally {
      isLoading = false;
    }
  }

  async function handleClassSelection(event) {
    const newClassId = parseInt(event.target.value, 10);
    selectedClassId = newClassId;
    // Reset states
    classDetails = null;
    errorMessage = '';
    successMessage = '';
    attendanceCount = 0;

    if (newClassId) {
      fetchClassDetails(newClassId);
    }
  }

  async function handleSubmit() {
    if (!selectedClassId || !classDetails) {
      errorMessage = 'クラスが選択されていません。';
      console.log("selectedClassId:", selectedClassId);
      console.log("classDetails:", classDetails);
      return;
    }

    isLoading = true;
    errorMessage = '';
    successMessage = '';

    if (attendanceCount > classDetails.studentCount) {
      errorMessage = `出席人数がクラスの総人数（${classDetails.studentCount}人）を超えています。`;
      isLoading = false;
      return;
    }

    try {
      const response = await fetch('/api/admin/attendance/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          class_id: parseInt(selectedClassId),
          attendance_count: attendanceCount,
        }),
      });

      const result = await response.json();

      if (response.ok) {
        successMessage = result.message || '出席を正常に登録しました。';
        // Refresh class details to show updated points
        await fetchClassDetails(selectedClassId);
      } else {
        errorMessage = result.error || '出席の登録に失敗しました。';
      }
    } catch (err) {
      errorMessage = '通信エラーが発生しました。後ほどもう一度お試しください。';
      console.error('Submission error:', err);
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="container mx-auto p-8">
  <h1 class="text-3xl font-bold mb-6">出席点管理</h1>

  <div class="bg-white shadow-md rounded-lg p-6">
    <!-- Conditional Class Display -->
    {#if managedClass}
      <div class="mb-6">
        <h2 class="text-xl font-semibold text-gray-800">対象クラス: {managedClass.name}</h2>
        <p class="text-sm text-gray-500">あなたの担当クラスが自動的に選択されています。</p>
      </div>
    {:else}
      <!-- Class Selector for non-reps -->
      <div class="mb-6">
        <label for="classSelector" class="block text-sm font-medium text-gray-700 mb-2">対象クラスを選択</label>
        <select
          id="classSelector"
          class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
          on:change={handleClassSelection}
          bind:value={selectedClassId}
        >
          <option value={null}>-- クラスを選択してください --</option>
          {#each classes as cls}
            <option value={cls.id}>{cls.name}</option>
          {/each}
        </select>
      </div>
    {/if}

    {#if isLoading && !classDetails}
      <p>クラス情報を読み込み中...</p>
    {/if}

    <!-- Attendance Form -->
    {#if classDetails}
      <div>
        <h2 class="text-2xl font-semibold mb-2">{classDetails.name}</h2>
        <div class="text-gray-600 mb-4">
          <p>クラスの総人数: {classDetails.studentCount}人</p>
          <p>現在の出席ポイント: {classDetails.attendancePoints}ポイント</p>
        </div>

        <form on:submit|preventDefault={handleSubmit}>
          <div class="mb-4">
            <label for="attendanceCount" class="block text-sm font-medium text-gray-700">出席人数</label>
            <input
              type="number"
              id="attendanceCount"
              bind:value={attendanceCount}
              min="0"
              max={classDetails.studentCount}
              required
              class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            />
          </div>

          {#if errorMessage}
            <p class="text-red-500 text-sm mb-4">{errorMessage}</p>
          {/if}

          {#if successMessage}
            <p class="text-green-500 text-sm mb-4">{successMessage}</p>
          {/if}

          <button
            type="submit"
            disabled={isLoading}
            class="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            {isLoading ? '登録中...' : '出席を登録する'}
          </button>
        </form>
      </div>
    {/if}
  </div>
</div>