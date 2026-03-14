import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, request, locals }) {
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
			notifications: notifications ?? [],
			user: locals.user
		};
	} catch (error) {
		console.error('Error loading notifications:', error);
		return {
			notifications: [],
			user: locals.user,
			error: error.message
		};
	}
}

/** @type {import('./$types').Actions} */
export const actions = {
	updateFilters: async ({ request, fetch }) => {
		const data = await request.formData();
		const filters = data.getAll('filters');

		const headers = {
			'Content-Type': 'application/json',
			cookie: request.headers.get('cookie')
		};
		const authHeader = request.headers.get('Authorization');
		if (authHeader) {
			headers.Authorization = authHeader;
		}

		try {
			const response = await fetch(`${BACKEND_URL}/api/notifications/filters`, {
				method: 'PUT',
				headers,
				body: JSON.stringify({ filters })
			});

			if (!response.ok) {
				const errorText = await response.text();
				return { error: `フィルタ更新に失敗しました: ${response.status} ${errorText}` };
			}

			const result = await response.json();
			return { message: result.message };
		} catch (error) {
			console.error('Error updating filters:', error);
			return { error: 'フィルタ更新中にエラーが発生しました' };
		}
	}
};

