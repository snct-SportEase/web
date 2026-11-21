import * as env from '$app/env/private';
import type { PageServerLoad } from './$types';
const BACKEND_URL = env.BACKEND_URL;

export const load = (async ({ fetch, request }: Pick<Parameters<PageServerLoad>[0], 'fetch' | 'request'>) => {
  try {
	const headers: Record<string, string> = { cookie: request.headers.get('cookie') ?? '' };
	const authorization = request.headers.get('authorization');
	if (authorization) headers.authorization = authorization;
    const response = await fetch(BACKEND_URL + `/api/classes`, { headers });
    if (!response.ok) {
      throw new Error(`Failed to fetch classes: ${response.statusText}`);
    }
    const fetchedClasses: { id: number; name: string; student_count?: number }[] = await response.json();
    const classes = fetchedClasses.filter((classItem) => classItem.name !== '専教');
    return {
      classes,
    };
  } catch (error) {
    console.error('Error loading classes:', error);
    return {
      classes: [],
      error: 'クラスの読み込みに失敗しました。',
    };
  }
}) satisfies PageServerLoad;
