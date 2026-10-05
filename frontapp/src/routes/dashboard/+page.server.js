import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { createBackendSessionHeaders } from '$lib/server/backendSessionHeaders.js';
const BACKEND_URL = env.BACKEND_URL;

/** @type {import('./$types').PageServerLoad} */
export async function load({ locals, fetch, request }) {
	const backendHeaders = {
		cookie: request.headers.get('cookie') ?? ''
	};
	const authorization = request.headers.get('authorization');
	if (authorization) backendHeaders.authorization = authorization;

  const returnData = {
    user: locals.user,
    classes: [],
    events: [],
    isClassMember: false,
    className: null,
    classInfo: null,
    members: [],
    progress: []
  };

  const tasks = [];
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
          const payload = await response.json();
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
}

/** @type {import('./$types').Actions} */
export const actions = {
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
