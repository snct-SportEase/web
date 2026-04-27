import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

const BACKEND_URL = env.BACKEND_URL;

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, locals, request }) {
  const user = locals.user;
  const isRoot = user?.roles?.some((role) => role.name === 'root') ?? false;
  const isAdmin = user?.roles?.some((role) => role.name === 'admin') ?? false;
  const classRole = user?.roles?.find(
    (role) => typeof role?.name === 'string' && role.name.endsWith('_rep')
  );
  const managedClassName = classRole?.name?.slice(0, -4) ?? null;

  const headers = {
    cookie: request.headers.get('cookie') ?? ''
  };

  const authHeader = request.headers.get('authorization');
  if (authHeader) {
    headers.authorization = authHeader;
  }

  try {
    const classesResponse = await fetch(`${BACKEND_URL}/api/classes`, { headers });
    if (!classesResponse.ok) {
      throw error(classesResponse.status, 'Failed to fetch classes.');
    }

    const classesPayload = await classesResponse.json();
    const allClasses = Array.isArray(classesPayload) ? classesPayload : [];

    if (isRoot) {
      return {
        classes: allClasses,
        managedClass: null,
        canSelectAllClasses: true,
        restrictionError: ''
      };
    }

    if (isAdmin) {
      const managedClass = allClasses.find((cls) => cls?.name === managedClassName) ?? null;

      return {
        classes: managedClass ? [managedClass] : [],
        managedClass,
        canSelectAllClasses: false,
        restrictionError: managedClass
          ? ''
          : '担当クラスが見つからないため、出席点管理を利用できません。'
      };
    }

    return {
      classes: allClasses,
      managedClass: null,
      canSelectAllClasses: false,
      restrictionError: ''
    };
  } catch (err) {
    if (err?.status) {
      throw err;
    }

    console.error('Unexpected error in attendance-management load:', err);
    throw error(500, 'An unexpected error occurred.');
  }
}
