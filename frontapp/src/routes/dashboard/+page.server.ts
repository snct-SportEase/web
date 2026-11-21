import { redirect } from '@sveltejs/kit';
import * as env from '$app/env/private';
import { createBackendSessionHeaders } from '$lib/server/backendSessionHeaders.js';
import type { Actions, PageServerLoad } from './$types';

type ClassInfo = { id: number; name: string; student_count: number; attend_count: number };
type ClassMember = { id: string; email: string; display_name?: string | null; assignments?: { sport_name: string; team_name: string }[] };
type ProgressMatch = { match_id: number; round: number; round_label: string; opponent_name?: string; match_status?: string; start_time?: string | null; result?: string; score?: string | null };
type ClassProgress = { sport_name: string; team_name: string; tournament_name: string; status: string; current_round: string; next_match?: ProgressMatch | null; last_match?: ProgressMatch | null };
type DashboardData = { user: App.Locals['user']; classes: ClassInfo[]; events: { id: number; name: string }[]; isClassMember: boolean; className: string | null; classInfo: ClassInfo | null; members: ClassMember[]; progress: ClassProgress[] };
const BACKEND_URL = env.BACKEND_URL;

export const load = (async ({ locals, fetch, request }: Pick<Parameters<PageServerLoad>[0], 'locals' | 'fetch' | 'request'>) => {
	const backendHeaders: Record<string, string> = {
		cookie: request.headers.get('cookie') ?? ''
	};
	const authorization = request.headers.get('authorization');
	if (authorization) backendHeaders.authorization = authorization;

  const returnData: DashboardData = {
    user: locals.user,
    classes: [],
    events: [],
    isClassMember: false,
    className: null,
    classInfo: null,
    members: [],
    progress: []
  };

  const tasks: (() => Promise<void>)[] = [];
  if (!locals.user?.is_profile_complete) {
    tasks.push(async () => {
      try {
        const response = await fetch(`${BACKEND_URL}/api/classes`, { headers: backendHeaders });
        if (response.ok) returnData.classes = await response.json();
      } catch (error) {
        console.error('Failed to fetch classes:', error);
      }
    });
  }

  const isRoot = locals.user?.roles?.some(role => role.name === 'root');
  if (isRoot && locals.user?.is_profile_complete) {
    tasks.push(async () => {
      try {
        const response = await fetch(`${BACKEND_URL}/api/root/events`, { headers: backendHeaders });
        if (response.ok) returnData.events = await response.json();
      } catch (error) {
        console.error('Failed to fetch events:', error);
      }
    });
  }

  if (locals.user?.class_id) {
    tasks.push(async () => {
      try {
        const response = await fetch(`${BACKEND_URL}/api/student/class-progress`, { headers: backendHeaders });
        if (response.ok) {
          const payload: { class_name?: string; class_info?: ClassInfo; members?: ClassMember[]; progress?: ClassProgress[] } = await response.json();
          returnData.isClassMember = true;
          returnData.className = payload.class_name ?? null;
          returnData.classInfo = payload.class_info ?? null;
          returnData.members = payload.members ?? [];
          returnData.progress = payload.progress ?? [];
        }
      } catch (error) {
        console.error('Failed to fetch class progress:', error);
      }
    });
  }
  await Promise.all(tasks.map(task => task()));

  return returnData;
}) satisfies PageServerLoad;

export const actions: Actions = {
  logout: async ({ fetch, locals, cookies }) => {
    await fetch(`${BACKEND_URL}/api/auth/logout`, {
      method: 'POST',
      headers: createBackendSessionHeaders(cookies)
    });

    // Clear the user from locals and redirect
    locals.user = null;
    cookies.delete('session_token', { path: '/' });
    cookies.delete('csrf_token', { path: '/' });
    throw redirect(302, '/');
  },
};
