import { redirect } from '@sveltejs/kit';

/** @type {import('./$types').LayoutServerLoad} */
export const load = async ({ parent }) => {
  // 親の+layout.server.jsからsession, userProfile, classesを取得
  const { session, userProfile, classes } = await parent();

  // セッションがない（未ログイン）場合、ホームページにリダイレクト
  if (!session) {
    throw redirect(303, '/');
  }

  // データを子コンポーネントに渡す
  return {
    session,
    userProfile,
    classes
  };
};
