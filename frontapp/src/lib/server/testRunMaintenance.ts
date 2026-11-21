const ISOLATING_STATES = new Set(['starting', 'testing', 'restoring', 'failed']);

type UserWithRoles = { roles?: { name: string }[] } | null | undefined;

export function isStudentOnly(user: UserWithRoles): boolean {
  const roles = user?.roles ?? [];
  const hasTestParticipantRole = roles.some((role) => role.name === 'admin' || role.name === 'root');
  return !hasTestParticipantRole && roles.some((role) => role.name === 'student');
}

export function requiresTestRunMaintenance(user: UserWithRoles, testRunState: string, stateAvailable = true): boolean {
  if (!isStudentOnly(user)) return false;
  if (!stateAvailable) return true;
  return ISOLATING_STATES.has(testRunState);
}
