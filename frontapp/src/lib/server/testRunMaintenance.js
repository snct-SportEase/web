const ISOLATING_STATES = new Set(['starting', 'testing', 'restoring', 'failed']);

export function isStudentOnly(user) {
  const roles = user?.roles ?? [];
  const hasTestParticipantRole = roles.some((role) => role.name === 'admin' || role.name === 'root');
  return !hasTestParticipantRole && roles.some((role) => role.name === 'student');
}

export function requiresTestRunMaintenance(user, testRunState, stateAvailable = true) {
  if (!isStudentOnly(user)) return false;
  if (!stateAvailable) return true;
  return ISOLATING_STATES.has(testRunState);
}
