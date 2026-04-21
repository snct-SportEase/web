import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter(),
		csrf: {
			trustedOrigins: ['*']
		},
		csp: {
			// mode: 'auto' でSvelteKitが生成するインラインスクリプト/スタイルに
			// 自動でnonceを付与する
			mode: 'auto',
			directives: {
				'default-src': ['self'],
				// script-src は mode:'auto' により 'nonce-{random}' が自動付与される
				'script-src': ['self'],
				// bracketry が描画時に <style> 注入と element.style 更新を行うため、
				// トーナメント表の描画ページでは inline style の許可が必要。
				'style-src': ['self', 'unsafe-inline'],
				// 画像: アップロード画像(同一オリジン) + data URI (Chart.js等)
				'img-src': ['self', 'data:', 'blob:'],
				// WebSocket接続は同一オリジン (wss/ws)
				'connect-src': ['self'],
				'font-src': ['self'],
				// frameとobjectは不要
				'frame-src': ['none'],
				'object-src': ['none'],
				// base タグによるURL書き換えを防止
				'base-uri': ['self'],
				// フォーム送信先を同一オリジンに限定
				'form-action': ['self'],
			}
		}
	}
};

export default config;
