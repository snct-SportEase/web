import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, request }) {
	const headers = {
		cookie: request.headers.get('cookie')
	};
	const authHeader = request.headers.get('Authorization');
	if (authHeader) {
		headers.Authorization = authHeader;
	}

	try {
		const params = new URLSearchParams({
			limit: '100'
		});
		const response = await fetch(`${BACKEND_URL}/api/notifications?${params.toString()}`, {
			headers
		});

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`Failed to fetch notifications: ${response.status} ${errorText}`);
		}

		const { notifications } = await response.json();
		return {
			notifications: notifications ?? []
		};
	} catch (error) {
		console.error('Error loading notifications:', error);
		return {
			notifications: [],
			error: error.message
		};
	}
}

