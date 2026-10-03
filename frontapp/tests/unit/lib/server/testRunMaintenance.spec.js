import { describe, expect, it } from 'vitest';
import { isStudentOnly, requiresTestRunMaintenance } from '$lib/server/testRunMaintenance.js';

describe('test run maintenance isolation', () => {
  const student = { roles: [{ name: 'student' }] };
  const admin = { roles: [{ name: 'admin' }] };
  const root = { roles: [{ name: 'root' }] };

  it('treats only student accounts as non-participants', () => {
    expect(isStudentOnly(student)).toBe(true);
    expect(isStudentOnly(admin)).toBe(false);
    expect(isStudentOnly(root)).toBe(false);
  });

  it.each(['starting', 'testing', 'restoring', 'failed'])('isolates students in %s state', (state) => {
    expect(requiresTestRunMaintenance(student, state)).toBe(true);
    expect(requiresTestRunMaintenance(admin, state)).toBe(false);
  });

  it('releases students while only notification delivery remains paused', () => {
    expect(requiresTestRunMaintenance(student, 'awaiting_notification_resume')).toBe(false);
  });

  it('fails closed when the state cannot be loaded', () => {
    expect(requiresTestRunMaintenance(student, '', false)).toBe(true);
  });
});
