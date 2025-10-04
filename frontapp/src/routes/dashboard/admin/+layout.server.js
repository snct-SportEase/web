import { redirect } from '@sveltejs/kit';

/** @type {import('./$types').LayoutServerLoad} */
export function load({ locals }) {
  const user = locals.user;

  if (!user || !user.roles.some(role => role.name === 'admin' || role.name === 'root')) {
    throw redirect(303, '/dashboard');
  }

  return {
    user: user,
  };
}
