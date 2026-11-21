import { error } from '@sveltejs/kit';
import * as env from '$app/env/private';
import type { PageServerLoad } from './$types';

const BACKEND_URL = env.BACKEND_URL;
type AttendanceClass = { id: number; name: string; student_count: number };
type AttendancePage = { classes: AttendanceClass[]; managedClass: AttendanceClass | null; canSelectAllClasses: boolean; restrictionError: string };

export const load: PageServerLoad = async ({ fetch, locals, request }): Promise<AttendancePage> => {
  const user = locals.user;
  const isRoot = user?.roles?.some((role) => role.name === 'root') ?? false;
  const isAdmin = user?.roles?.some((role) => role.name === 'admin') ?? false;

  const headers: Record<string, string> = {
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

    if (isRoot || isAdmin) {
      return {
        classes: allClasses,
        managedClass: null,
        canSelectAllClasses: true,
        restrictionError: ''
      };
    }

    return {
      classes: allClasses,
      managedClass: null,
      canSelectAllClasses: false,
      restrictionError: ''
    };
  } catch (err) {
    if (typeof err === 'object' && err !== null && 'status' in err && err.status) {
      throw err;
    }

    console.error('Unexpected error in attendance-management load:', err);
    throw error(500, 'An unexpected error occurred.');
  }
};
