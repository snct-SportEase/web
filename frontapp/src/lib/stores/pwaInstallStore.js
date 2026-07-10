import { browser } from '$app/environment';
import { writable } from 'svelte/store';

export const pwaInstallPromptAvailable = writable(false);
export const pwaInstallDialogOpen = writable(false);

let deferredInstallPrompt = null;

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

function captureInstallPrompt(event) {
	event.preventDefault();
	deferredInstallPrompt = event;
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
