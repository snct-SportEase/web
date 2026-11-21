const RESULT_ENTRY_STATUSES = new Set(['active', 'testing']);

export function canEnterEventResults(status: string | null | undefined): boolean {
  return status != null && RESULT_ENTRY_STATUSES.has(status);
}
