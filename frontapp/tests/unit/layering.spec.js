import { readFileSync, readdirSync } from 'node:fs';
import { extname, join } from 'node:path';
import { describe, expect, it } from 'vitest';

const projectRoot = new URL('../..', import.meta.url);
const sourceRoot = new URL('./src/', projectRoot);

function readLayerValue(css, name) {
	const match = css.match(new RegExp(`--app-layer-${name}:\\s*(\\d+);`));
	expect(match, `missing --app-layer-${name}`).not.toBeNull();
	return Number(match[1]);
}

function collectSvelteFiles(directory) {
	return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
		const path = join(directory, entry.name);
		if (entry.isDirectory()) return collectSvelteFiles(path);
		return extname(path) === '.svelte' ? [path] : [];
	});
}

describe('application layer policy', () => {
	it('keeps blocking overlays above the header and passive notifications', () => {
		const css = readFileSync(new URL('./src/app.css', projectRoot), 'utf8');
		const header = readLayerValue(css, 'header');
		const notification = readLayerValue(css, 'notification');
		const modal = readLayerValue(css, 'modal');
		const raisedModal = readLayerValue(css, 'modal-raised');

		expect(notification).toBeGreaterThan(header);
		expect(modal).toBeGreaterThan(notification);
		expect(raisedModal).toBeGreaterThan(modal);
	});

	it('requires every viewport-fixed Svelte element to use a semantic layer', () => {
		const violations = [];
		const fixedClassPattern = /<[^>]+class="([^"]*\bfixed\b[^"]*)"[^>]*>/gs;

		for (const file of collectSvelteFiles(sourceRoot.pathname)) {
			const source = readFileSync(file, 'utf8');
			for (const match of source.matchAll(fixedClassPattern)) {
				if (!match[1].includes('app-layer-')) {
					violations.push(`${file.replace(sourceRoot.pathname, '')}: ${match[1]}`);
				}
			}
		}

		expect(violations).toEqual([]);
	});
});
