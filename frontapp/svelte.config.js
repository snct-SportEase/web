import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
    kit: {
        adapter: adapter({
            // デフォルトの出力ディレクトリは 'build' です
            pages: 'build',
            assets: 'build',
            fallback: 'index.html', // SPAの場合に重要
            precompress: false,
            strict: true
        })
    }
};

export default config;
