import * as env from '$app/env/private';
import { createBackendSessionHeaders } from '$lib/server/backendSessionHeaders.js';
import type { Actions, PageServerLoad } from './$types';
import type { Notification } from '$lib/types/notification';
const BACKEND_URL = env.BACKEND_URL;

export const load: PageServerLoad = async ({ fetch, request, locals }) => {
	const headers: Record<string, string> = {
		cookie: request.headers.get('cookie') ?? ''
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

		const { notifications }: { notifications?: Notification[] } = await response.json();
		return {
			notifications: notifications ?? [],
			user: locals.user
		};
	} catch (error) {
		console.error('Error loading notifications:', error);
		return {
			notifications: [],
			user: locals.user,
			error: error instanceof Error ? error.message : String(error)
		};
	}
};

export const actions: Actions = {
	updateFilters: async ({ request, fetch, cookies }) => {
		const data = await request.formData();
		const filters = data.getAll('filters');

		const headers = createBackendSessionHeaders(cookies, {
			'Content-Type': 'application/json'
		});
		const authHeader = request.headers.get('Authorization');
		if (authHeader) {
			headers.set('Authorization', authHeader);
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

			const result: { message: string } = await response.json();
			return { message: result.message };
		} catch (error) {
			console.error('Error updating filters:', error);
			return { error: 'フィルタ更新中にエラーが発生しました' };
		}
	}
};
