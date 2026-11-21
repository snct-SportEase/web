import { browser } from '$app/env';
import { writable } from 'svelte/store';

export const pwaInstallPromptAvailable = writable(false);
export const pwaInstallDialogOpen = writable(false);

interface InstallPromptEvent extends Event {
	prompt(): Promise<{ outcome: 'accepted' | 'dismissed'; platform?: string }>;
}
let deferredInstallPrompt: InstallPromptEvent | null = null;

export function openPWAInstallDialog() {
	pwaInstallDialogOpen.set(true);
}

export function closePWAInstallDialog() {
	pwaInstallDialogOpen.set(false);
}

export async function promptPWAInstall() {
	if (!deferredInstallPrompt || typeof deferredInstallPrompt.prompt !== 'function') {
		return { outcome: 'unavailable' };
	}

	const installPrompt = deferredInstallPrompt;
	deferredInstallPrompt = null;
	pwaInstallPromptAvailable.set(false);

	const result = await installPrompt.prompt();
	if (result?.outcome === 'accepted') {
		closePWAInstallDialog();
	}

	return result;
}

function captureInstallPrompt(event: Event) {
	event.preventDefault();
	deferredInstallPrompt = event as InstallPromptEvent;
	pwaInstallPromptAvailable.set(true);
}

function handleAppInstalled() {
	deferredInstallPrompt = null;
	pwaInstallPromptAvailable.set(false);
	closePWAInstallDialog();
}

if (browser) {
	window.addEventListener('beforeinstallprompt', captureInstallPrompt);
	window.addEventListener('appinstalled', handleAppInstalled);
}
