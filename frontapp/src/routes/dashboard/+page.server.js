import { fail, redirect } from '@sveltejs/kit';
import { BACKEND_URL } from '$env/static/private';

/** @type {import('./$types').PageServerLoad} */
export async function load({ locals, fetch, request }) {
  const returnData = {
    user: locals.user,
    classes: [],
    events: [],
    isClassRep: false,
    className: null,
    classInfo: null,
    members: [],
    progress: []
  };

  try {
    // Fetch classes
    const classesResponse = await fetch(`${BACKEND_URL}/api/classes`);
    if (classesResponse.ok) {
      returnData.classes = await classesResponse.json();
    }
  } catch (e) {
    console.error('Failed to fetch classes:', e);
  }

  // Fetch events if user is root
  const isRoot = locals.user?.roles?.some(role => role.name === 'root');
  if (isRoot && locals.user?.is_profile_complete) {
      try {
        const eventResponse = await fetch(`${BACKEND_URL}/api/root/events`, {
            headers: {
                'cookie': request.headers.get('cookie'),
            }
        });
        if (eventResponse.ok) {
            returnData.events = await eventResponse.json();
        }
      } catch (e) {
        console.error('Failed to fetch events:', e);
      }
  }

  const classRole = locals.user?.roles?.find(role => typeof role.name === 'string' && role.name.endsWith('_rep'));
  if (classRole) {
    try {
      const response = await fetch(`${BACKEND_URL}/api/student/class-progress`, {
        headers: {
          cookie: request.headers.get('cookie')
        }
      });
      if (response.ok) {
        const payload = await response.json();
        returnData.isClassRep = true;
        returnData.className = payload.class_name ?? classRole.name.replace(/_rep$/, '');
        returnData.classInfo = payload.class_info ?? null;
        returnData.members = payload.members ?? [];
        returnData.progress = payload.progress ?? [];
      } else if (response.status === 403) {
        returnData.isClassRep = false;
      }
    } catch (e) {
      console.error('Failed to fetch class progress:', e);
    }
  }

  return returnData;
}

/** @type {import('./$types').Actions} */
export const actions = {
  logout: async ({ fetch, locals, request }) => {
    const sessionCookie = request.headers.get('cookie');
    await fetch(`${BACKEND_URL}/api/auth/logout`, {
        method: 'POST',
        headers: {
            'cookie': sessionCookie,
        },
    });

    // Clear the user from locals and redirect
    locals.user = null;
    throw redirect(302, '/');
  },
};