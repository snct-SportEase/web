const RESULT_ENTRY_STATUSES = new Set(['active', 'testing']);

export function canEnterEventResults(status) {
  return RESULT_ENTRY_STATUSES.has(status);
}
