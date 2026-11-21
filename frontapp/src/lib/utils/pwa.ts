import { browser } from '$app/env';

/**
 * PWAがインストールされているか（standaloneモード）を検出
 */
export function isPWAInstalled(): boolean {
	if (!browser) return false;
	
	// standaloneモードで実行されているか
	if (window.matchMedia('(display-mode: standalone)').matches) {
		return true;
	}
	
	// iOS Safariの場合の検出
	if ((window.navigator as Navigator & { standalone?: boolean }).standalone === true) {
		return true;
	}
	
	return false;
}

/**
 * PWAがインストール可能かどうかを検出
 */
export function isPWAInstallable(): boolean {
	if (!browser) return false;
	
	// Service Workerがサポートされているか
	if (!('serviceWorker' in navigator)) {
		return false;
	}
	
	// 既にインストール済みの場合はfalse
	if (isPWAInstalled()) {
		return false;
	}
	
	return true;
}

/**
 * 現在のOS/デバイスタイプを検出
 */
export function getDeviceType(): 'ios' | 'android' | 'windows' | 'macos' | 'linux' | 'unknown' {
	if (!browser) return 'unknown';
	
	const userAgent = navigator.userAgent || navigator.vendor || (window as Window & { opera?: string }).opera || '';
	
	// iOS
	if (/iPad|iPhone|iPod/.test(userAgent) && !(window as Window & { MSStream?: unknown }).MSStream) {
		return 'ios';
	}
	
	// Android
	if (/android/i.test(userAgent)) {
		return 'android';
	}
	
	// Windows
	if (/Windows/.test(userAgent)) {
		return 'windows';
	}
	
	// macOS
	if (/Macintosh|Mac OS X/.test(userAgent)) {
		return 'macos';
	}
	
	// Linux
	if (/Linux/.test(userAgent)) {
		return 'linux';
	}
	
	return 'unknown';
}

/**
 * ブラウザタイプを検出
 */
export function getBrowserType(): 'chrome' | 'edge' | 'safari' | 'firefox' | 'unknown' {
	if (!browser) return 'unknown';
	
	const userAgent = navigator.userAgent || navigator.vendor || (window as Window & { opera?: string }).opera || '';
	
	// Chrome
	if (/Chrome/.test(userAgent) && !/Edg/.test(userAgent)) {
		return 'chrome';
	}
	
	// Edge
	if (/Edg/.test(userAgent)) {
		return 'edge';
	}
	
	// Safari
	if (/Safari/.test(userAgent) && !/Chrome/.test(userAgent)) {
		return 'safari';
	}
	
	// Firefox
	if (/Firefox/.test(userAgent)) {
		return 'firefox';
	}
	
	return 'unknown';
}
