import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

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

		const activeEventResponse = await fetch(`${BACKEND_URL}/api/events/active`, {
			headers
		});
		if (activeEventResponse.ok) {
			const activeEvent = await activeEventResponse.json();
			if (activeEvent?.hide_scores) {
				return { scores: [], error: '得点一覧は現在非表示です。' };
			}
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
