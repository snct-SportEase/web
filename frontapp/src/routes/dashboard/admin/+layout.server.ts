import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = ({ locals }) => {
  const user = locals.user;

  if (!user || !user.roles?.some(role => role.name === 'admin' || role.name === 'root')) {
    throw redirect(303, '/dashboard');
  }

  return {
    user: user,
  };
};
