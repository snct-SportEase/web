import { normalizeActiveEvent } from '$lib/utils/activeEvent.js';

/** @type {import('./$types').LayoutServerLoad} */
export function load({ locals, url }) {
  // Track route changes so the browser receives the hook's latest snapshot.
  void url.pathname;
  return {
    user: locals.user,
    activeEvent: normalizeActiveEvent(locals.activeEvent),
    activeEventAvailable: locals.activeEvent !== null
  };
}
