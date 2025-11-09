import { BACKEND_URL } from '$env/static/private';

/** @type {import('./$types').PageServerLoad} */
export async function load({ locals, fetch, request }) {
	const user = locals.user;

	if (!user) {
		return { user: null, isAdmin: false };
	}

	// Check if user is admin or root
	const isAdmin = user.roles?.some((role) => role.name === 'admin' || role.name === 'root') || false;

	return {
		user: user,
		isAdmin: isAdmin
	};
}

