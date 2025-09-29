import { createClient } from '@supabase/supabase-js';
import { PUBLIC_SUPABASE_URL } from '$env/static/public';
import { SUPABASE_SERVICE_KEY } from '$env/static/private';

// 管理者権限を持つSupabaseクライアントを初期化
const supabaseAdmin = createClient(PUBLIC_SUPABASE_URL, SUPABASE_SERVICE_KEY);

/** @type {import('./$types').LayoutServerLoad} */
export const load = async ({ locals: { getUser } }) => {
  const user = await getUser();
  let userProfile = null;
  let classes = [];
  console.log('user', user);

  if (user) {
    // ログインしているユーザーのプロフィールを単純に取得する
    const { data: profileData, error: profileError } = await supabaseAdmin
      .from('users')
      .select(`*, class:classes!class_id (name)`)
      .eq('id', user.id)
      .single();

    // もしここでエラーが出た場合、それはトリガーが失敗しているなど重大な問題
    if (profileError) {
      console.error(
        'Critical Error: User is authenticated, but their profile is missing. Check the database trigger.',
        profileError
      );
    } else {
      userProfile = profileData;
    }

    // クラスリストを取得
    const { data: classesData, error: classesError } = await supabaseAdmin
      .from('classes')
      .select('id, name')
      .order('name');

    if (classesError) {
      console.error('Error fetching classes:', classesError.message);
    } else {
      classes = classesData || [];
    }
  }

  return {
    user,
    userProfile,
    classes
  };
};