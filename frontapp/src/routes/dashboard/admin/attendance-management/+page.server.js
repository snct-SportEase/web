import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

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

    // Check if user has a class_name_rep role
    if (user && user.roles) {
      const classRole = user.roles.find(
        (role) => typeof role?.name === 'string' && role.name.endsWith('_rep')
      );

      if (classRole) {
        const className = classRole.name.slice(0, -4); // Remove '_rep' suffix
        const classDetailsResponse = await fetch(`${BACKEND_URL}/api/class/name/${className}`);

        if (classDetailsResponse.ok) {
          managedClass = await classDetailsResponse.json();
        } else {
          // It's okay if this fails, the user can still select manually
          console.warn(`Could not automatically fetch class details for role ${classRole.name}`);
        }
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
