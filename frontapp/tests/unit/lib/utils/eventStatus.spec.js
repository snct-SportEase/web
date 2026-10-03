import { describe, expect, it } from 'vitest';
import { canEnterEventResults } from '$lib/utils/eventStatus.js';

describe('canEnterEventResults', () => {
  it.each(['active', 'testing'])('%s では結果を入力できる', (status) => {
    expect(canEnterEventResults(status)).toBe(true);
  });

  it.each(['preparing', 'upcoming', 'archived', '', undefined])(
    '%s では結果を入力できない',
    (status) => {
      expect(canEnterEventResults(status)).toBe(false);
    }
  );
});
