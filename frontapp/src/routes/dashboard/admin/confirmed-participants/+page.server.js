import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, request, locals }) {
	const user = locals.user;
	if (!user) {
		return { classes: [] };
	}

	const headers = {
		cookie: request.headers.get('cookie')
	};
	const authHeader = request.headers.get('Authorization');
	if (authHeader) {
		headers.Authorization = authHeader;
	}

	let classes = [];
	try {
		const response = await fetch(`${BACKEND_URL}/api/classes`, { headers });
		if (response.ok) {
			classes = await response.json();
		}
	} catch (error) {
		console.error('Error loading classes:', error);
	}

	return { classes };
}

