import type { BracketOptions } from '@snct-sportease/matchline';

/** ポインターを持たない端末ではタップで勝ち上がりを確認できるようにする。 */
export function getMatchlineOptions(): BracketOptions {
  const canHover = typeof window === 'undefined' || typeof window.matchMedia !== 'function'
    || window.matchMedia('(hover: hover)').matches;

  return {
    connectorType: 'orthogonal',
    highlightTrigger: canHover ? 'hover' : 'click'
  };
}
