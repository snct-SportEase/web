import * as env from '$app/env/private';
import type { PageServerLoad } from './$types';
type Role = { name?: string; Name?: string; id?: number; ID?: number };
type Notification = { id: number; title: string; body: string; type: string; created_at: string; scheduled_at?: string | null; sent_at?: string | null; target_roles: string[]; target_user_count: number };
type SubscriptionStats = { target_user_count: number; subscribed_user_count: number; subscription_endpoint_count: number };
const BACKEND_URL = env.BACKEND_URL;

export const load: PageServerLoad = async ({ fetch, request }) => {
	const headers: Record<string, string> = {
		cookie: request.headers.get('cookie') ?? ''
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

		const notificationsPayload: { notifications?: Notification[] } = await notificationsResponse.json();
		let roles: Role[] = [];
		if (rolesResponse.ok) {
			const rolesPayload: { roles?: Role[] } = await rolesResponse.json();
			roles = rolesPayload.roles ?? [];
		} else {
			const errorText = await rolesResponse.text();
			console.error('Failed to fetch roles:', rolesResponse.status, errorText);
		}

		let subscriptionStats: SubscriptionStats | null = null;
		const defaultTargetRoles = roles.some((role) => (role.name ?? role.Name) === 'student')
			? ['student']
			: roles.slice(0, 1).map((role) => role.name ?? role.Name).filter((name): name is string => Boolean(name));
		if (defaultTargetRoles.length > 0) {
			const statsParams = new URLSearchParams();
			for (const role of defaultTargetRoles) {
				statsParams.append('roles', role);
			}

			const statsResponse = await fetch(`${BACKEND_URL}/api/root/notifications/subscription-stats?${statsParams.toString()}`, {
				headers
			});
			if (statsResponse.ok) {
				const statsPayload: { stats?: SubscriptionStats } = await statsResponse.json();
				subscriptionStats = statsPayload.stats ?? null;
			} else {
				const errorText = await statsResponse.text();
				console.error('Failed to fetch subscription stats:', statsResponse.status, errorText);
			}
		}

		return {
			notifications: notificationsPayload.notifications ?? [],
			roles,
			subscriptionStats
		};
	} catch (error) {
		console.error('Error loading notifications:', error);
		return {
			notifications: [],
			roles: [],
			subscriptionStats: null,
			error: error instanceof Error ? error.message : String(error)
		};
	}
};

