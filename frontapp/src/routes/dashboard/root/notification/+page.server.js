import { BACKEND_URL } from '$env/static/private';

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
			include_authored: 'true',
			limit: '100'
		});

		const [notificationsResponse, rolesResponse] = await Promise.all([
			fetch(`${BACKEND_URL}/api/notifications?${params.toString()}`, {
				headers
			}),
			fetch(`${BACKEND_URL}/api/root/notifications/roles`, {
				headers
			})
		]);

		if (!notificationsResponse.ok) {
			const errorText = await notificationsResponse.text();
			throw new Error(`Failed to fetch notifications: ${notificationsResponse.status} ${errorText}`);
		}

		const notificationsPayload = await notificationsResponse.json();
		let roles = [];
		if (rolesResponse.ok) {
			const rolesPayload = await rolesResponse.json();
			roles = rolesPayload.roles ?? [];
		} else {
			const errorText = await rolesResponse.text();
			console.error('Failed to fetch roles:', rolesResponse.status, errorText);
		}

		return {
			notifications: notificationsPayload.notifications ?? [],
			roles
		};
	} catch (error) {
		console.error('Error loading notifications:', error);
		return {
			notifications: [],
			roles: [],
			error: error.message
		};
	}
}


