import { get } from 'svelte/store';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
	closePWAInstallDialog,
	openPWAInstallDialog,
	promptPWAInstall,
	pwaInstallDialogOpen,
	pwaInstallPromptAvailable
} from '$src/lib/stores/pwaInstallStore.js';

describe('PWA install store', () => {
	beforeEach(() => {
		closePWAInstallDialog();
	});

	it('beforeinstallpromptを保持してユーザー操作時にブラウザのpromptを呼ぶ', async () => {
		const prompt = vi.fn(async () => ({ outcome: 'accepted', platform: 'web' }));
		const event = new Event('beforeinstallprompt', { cancelable: true });
		Object.defineProperty(event, 'prompt', { value: prompt });

		window.dispatchEvent(event);

		expect(event.defaultPrevented).toBe(true);
		expect(get(pwaInstallPromptAvailable)).toBe(true);

		openPWAInstallDialog();
		const result = await promptPWAInstall();

		expect(prompt).toHaveBeenCalledOnce();
		expect(result.outcome).toBe('accepted');
		expect(get(pwaInstallPromptAvailable)).toBe(false);
		expect(get(pwaInstallDialogOpen)).toBe(false);
	});

	it('利用可能なブラウザプロンプトがなければunavailableを返す', async () => {
		await expect(promptPWAInstall()).resolves.toEqual({ outcome: 'unavailable' });
	});
});
