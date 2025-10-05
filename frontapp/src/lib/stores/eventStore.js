import { writable, get } from 'svelte/store';

// activeEvent store holds the active event object or null
const { subscribe, set, update } = writable(null);

export const activeEvent = {
    subscribe,
    // internal setter
    _set: set,
    // initialize store by fetching current active event from backend
    init: async () => {
        try {
            const res = await fetch('/api/events/active');
            if (!res.ok) {
                console.warn('Failed to fetch active event:', res.status);
                set(null);
                return null;
            }
            const data = await res.json();
            // Expecting { event_id: <id> } from backend; fetch full event if id present
            if (data && data.event_id) {
                const evRes = await fetch(`/api/root/events`);
                if (evRes.ok) {
                    const events = await evRes.json();
                    const active = events.find(e => e.id === data.event_id) || null;
                    set(active);
                    return active;
                }
            }
            set(null);
            return null;
        } catch (err) {
            console.error('Error initializing activeEvent:', err);
            set(null);
            return null;
        }
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
                set(eventObj);
            } else {
                // clear
                await fetch('/api/root/events/active', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ event_id: null }),
                });
                set(null);
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
