<script>
    import { createEventDispatcher } from 'svelte';

    export let showModal = false;
    export let user = null;

    const dispatch = createEventDispatcher();
    const defaultRoles = ['root', 'admin', 'student'];

    function closeModal() {
        showModal = false;
    }

    async function deleteRole(roleName) {
        if (defaultRoles.includes(roleName)) {
            alert('デフォルトのロールは削除できません。');
            return;
        }

        if (!confirm(`${user.email}からロール「${roleName}」を削除しますか？`)) {
            return;
        }

        const res = await fetch('/api/admin/users/role', {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ user_id: user.id, role: roleName })
        });

        if (res.ok) {
            alert('ロールが削除されました。');
            dispatch('roleDeleted');
            closeModal();
        } else {
            const error = await res.json();
            alert(`ロールの削除に失敗しました: ${error.message}`);
        }
    }
</script>

{#if showModal && user}
<div class="fixed inset-0 bg-black bg-opacity-50 z-50 flex justify-center items-center">
    <div class="bg-white rounded-lg shadow-xl p-6 w-full max-w-md">
        <h2 class="text-xl font-bold mb-4">ロールの編集</h2>
        <p class="mb-2"><strong>ユーザー:</strong> {user.email}</p>
        
        <div class="mb-4">
            <h3 class="font-semibold">現在のロール:</h3>
            <div class="mt-2 space-y-2">
                {#each user.roles.filter(r => r.name !== 'student') as role}
                    <div class="flex items-center justify-between bg-gray-100 p-2 rounded-md">
                        <span class="text-sm font-medium text-gray-800">{role.name}</span>
                        {#if !defaultRoles.includes(role.name)}
                            <button on:click={() => deleteRole(role.name)} class="text-red-500 hover:text-red-700 font-semibold text-sm">
                                削除
                            </button>
                        {/if}
                    </div>
                {/each}
            </div>
        </div>

        <div class="flex justify-end">
            <button on:click={closeModal} class="px-4 py-2 bg-gray-200 text-gray-800 rounded-md hover:bg-gray-300">
                閉じる
            </button>
        </div>
    </div>
</div>
{/if}
