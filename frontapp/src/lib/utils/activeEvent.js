// The active endpoint includes all fields used by dashboard pages.
export function normalizeActiveEvent(payload) {
    if (!payload?.event_id) return null;
    return { ...payload, id: payload.event_id, name: payload.event_name };
}
