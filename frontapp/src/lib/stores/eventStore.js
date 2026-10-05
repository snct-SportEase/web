import { writable, get, derived } from 'svelte/store';
import { normalizeActiveEvent } from '$lib/utils/activeEvent.js';

// activeEvent store holds the active event object or null
const { subscribe, set } = writable(null);
let pending = null;
let loaded = false;
let loadedAt = 0;
let revision = 0;

export const activeEvent = {
    subscribe,
    // internal setter
    _set: (event) => activeEvent.seed(event),
    // Seed only in the browser; server loaders must keep data request-local.
    seed: (event) => {
        revision += 1;
        set(event);
        loaded = true;
        loadedAt = Date.now();
    },
    init: ({ force = false } = {}) => {
        if (pending) return pending;
        if (!force && loaded && Date.now() - loadedAt < 15_000) {
            return Promise.resolve(get({ subscribe }));
        }
        const requestRevision = revision;
        pending = (async () => {
            try {
                const res = await fetch('/api/events/active');
                if (!res.ok) throw new Error(`Failed to fetch active event: ${res.status}`);
                const event = normalizeActiveEvent(await res.json());
                if (requestRevision === revision) activeEvent.seed(event);
                return get({ subscribe });
            } catch (err) {
                console.error('Error initializing activeEvent:', err);
                if (requestRevision === revision) {
                    set(null);
                    loaded = false;
                }
                return get({ subscribe });
            } finally {
                pending = null;
            }
        })();
        return pending;
    },
    // set active event by passing full event object
    setActiveEvent: async (eventObj) => {
        try {
            if (eventObj && eventObj.id) {
                const res = await fetch('/api/root/events/active', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ event_id: eventObj.id }),
                });
                if (!res.ok) {
                    throw new Error('Failed to set active event on server');
                }
                activeEvent.seed(eventObj);
            } else {
                // clear
                const res = await fetch('/api/root/events/active', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ event_id: null }),
                });
                if (!res.ok) throw new Error('Failed to clear active event on server');
                activeEvent.seed(null);
            }
        } catch (err) {
            console.error('Failed to persist activeEvent to backend:', err);
            // do not update store on failure
            throw err;
        }
    },
    // set active event by id (helper)
    setActiveEventById: async (id) => {
        if (!id) {
            return activeEvent.setActiveEvent(null);
        }
        try {
            // fetch event details
            const res = await fetch('/api/root/events');
            if (!res.ok) throw new Error('Failed to fetch events');
            const events = await res.json();
            const eventObj = events.find(e => e.id === parseInt(id));
            if (!eventObj) throw new Error('Event not found');
            return activeEvent.setActiveEvent(eventObj);
        } catch (err) {
            console.error('Failed to set active event by id:', err);
            throw err;
        }
    },
    // read-only access to current value
    get: () => get({ subscribe }),
};

export const activeEventId = derived(
    activeEvent,
    $activeEvent => $activeEvent ? $activeEvent.id : null
);
