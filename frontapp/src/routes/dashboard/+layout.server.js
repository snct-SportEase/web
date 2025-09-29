/** @type {import('./$types').LayoutServerLoad} */
export const load = async ({ parent }) => {
  // 親の+layout.server.jsからuser, userProfile, classesを取得
  const { user, userProfile, classes } = await parent();

  // 認証チェックはhooks.server.jsに集約されているため、ここでのリダイレクトは不要

  // データを子コンポーネントに渡す
  return {
    user,
    userProfile,
    classes
  };
};