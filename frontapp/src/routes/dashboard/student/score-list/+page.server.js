import { redirect } from '@sveltejs/kit';
import { BACKEND_URL } from '$env/static/private';

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, locals, request }) {
	if (!locals.user) {
		throw redirect(302, '/');
	}

	try {
		const headers = {
			cookie: request.headers.get('cookie')
		};
		const authHeader = request.headers.get('Authorization');
		if (authHeader) {
			headers.Authorization = authHeader;
		}

		const response = await fetch(`${BACKEND_URL}/api/scores/class`, {
			headers
		});

		if (response.ok) {
			const scores = await response.json();
			return { scores };
		}

		return { scores: [] };
	} catch (error) {
		console.error('Error loading scores:', error);
		return { scores: [], error: error.message };
	}
}
