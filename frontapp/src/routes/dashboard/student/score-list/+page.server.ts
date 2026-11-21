import { redirect } from '@sveltejs/kit';
import * as env from '$app/env/private';
import type { PageServerLoad } from './$types';
import type { ClassScore } from '$lib/utils/sportScores';
export type RankedScore = ClassScore & { id: number; class_id: number; class_name: string; season: string; rank_current_event: number | null; rank_overall: number | null; total_points_current_event: number; total_points_overall: number; sport_names?: Record<string, string> };
const BACKEND_URL = env.BACKEND_URL;

export const load = (async ({ fetch, locals, request }: Pick<Parameters<PageServerLoad>[0], 'fetch' | 'request' | 'locals'>) => {
	if (!locals.user) {
		throw redirect(302, '/');
	}

	try {
		const headers: Record<string, string> = {
			cookie: request.headers.get('cookie') ?? ''
		};
		const authHeader = request.headers.get('Authorization');
		if (authHeader) {
			headers.Authorization = authHeader;
		}

		const response = await fetch(`${BACKEND_URL}/api/scores/class`, {
			headers
		});

		if (response.ok) {
			const scores: RankedScore[] = await response.json();
			return { scores };
		}

        if (response.status === 403) {
            return { scores: [], error: '得点一覧は現在非表示です。' };
        }
		return { scores: [] };
	} catch (error) {
		console.error('Error loading scores:', error);
		return { scores: [], error: error instanceof Error ? error.message : String(error) };
	}
}) satisfies PageServerLoad;
