import { fail, redirect } from '@sveltejs/kit';

/** @type {import('./$types').Actions} */
export const actions = {
  updateProfile: async ({ request, locals: { supabase, getUser } }) => {
    const user = await getUser();
    if (!user) {
      return fail(401, { message: 'Unauthorized' });
    }

    const formData = await request.formData();
    const displayName = formData.get('displayName');
    const classId = formData.get('classId');

    if (!displayName || !classId) {
      return fail(400, { message: '表示名とクラスの両方を選択してください。' });
    }

    const { error } = await supabase
      .from('users')
      .update({
        display_name: displayName,
        class_id: classId,
        is_profile_complete: true
      })
      .eq('id', user.id);

    if (error) {
      console.error('Error updating profile:', error);
      return fail(500, { message: 'プロフィールの更新に失敗しました。' });
    }

    console.log('Profile updated successfully');

    // 成功した場合、ダッシュボードにリダイレクトする
    throw redirect(303, '/dashboard');
  }
};
