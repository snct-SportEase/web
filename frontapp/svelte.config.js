import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter(),
		csp: {
			// mode: 'auto' でSvelteKitが生成するインラインスクリプト/スタイルに
			// 自動でnonceを付与する
			mode: 'auto',
			directives: {
				'default-src': ['self'],
				// script-src は mode:'auto' により 'nonce-{random}' が自動付与される
				'script-src': ['self'],
				// style-src も nonce のみ。インラインスタイルは <style> ブロックへ移行済み。
				// マークダウンの色付けは data-mk-color 属性 + JS 適用で実現。
				'style-src': ['self'],
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
