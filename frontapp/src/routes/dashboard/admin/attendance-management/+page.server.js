import { error } from '@sveltejs/kit';
import { BACKEND_URL } from '$env/static/private';

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, locals }) {
  try {
    // Fetch all classes for the selector
    const classesResponse = await fetch(BACKEND_URL + '/api/classes');
    if (!classesResponse.ok) {
      throw error(classesResponse.status, 'Failed to fetch classes.');
    }
    const classes = await classesResponse.json();

    let managedClass = null;
    const user = locals.user;

    if (user && user.role && user.role.endsWith('_rep')) {
      const className = user.role.replace('_rep', '');
      const classDetailsResponse = await fetch(`${BACKEND_URL}/api/class/name/${className}`);

      if (classDetailsResponse.ok) {
        managedClass = await classDetailsResponse.json();
      } else {
        // It's okay if this fails, the user can still select manually
        console.warn(`Could not automatically fetch class details for role ${user.role}`);
      }
    }

    return {
      classes,
      managedClass,
    };
  } catch (err) {
    if (err.status) {
      throw err;
    }
    console.error('Unexpected error in load function:', err);
    throw error(500, 'An unexpected error occurred.');
  }
}
