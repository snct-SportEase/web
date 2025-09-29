import { fail, json, redirect } from '@sveltejs/kit';

/** @type {import('./$types').RequestHandler} */
export async function POST({ request, locals: { supabase, getUser } }) {
  const user = await getUser();
  if (!user) {
    return json({ message: 'Unauthorized' }, { status: 401 });
  }

  const formData = await request.formData();
  const displayName = formData.get('displayName');
  const classId = formData.get('classId');

  // バリデーション
  if (!displayName || !classId) {
    return json({ message: '表示名とクラスの両方を選択してください。' }, { status: 400 });
  }

  // Supabaseのusersテーブルを更新
  const { error } = await supabase
    .from('users')
    .update({
      display_name: displayName,
      class_id: classId,
      is_profile_complete: true // プロフィールが完成したことをマーク
    })
    .eq('id', user.id);

  if (error) {
    console.error('Error updating profile:', error);
    return json({ message: 'プロフィールの更新に失敗しました。' }, { status: 500 });
  }

  // 更新が成功したら、成功ステータスを返す
  return json({ success: true }, { status: 200 });
}