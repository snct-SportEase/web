// For more info, see https://github.com/storybookjs/eslint-plugin-storybook#configuration-flat-config-format
import storybook from "eslint-plugin-storybook";

import { fileURLToPath } from 'node:url';
import { includeIgnoreFile } from '@eslint/compat';
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import svelteConfig from './sveltekit.config.ts';

const gitignorePath = fileURLToPath(new URL('./.gitignore', import.meta.url));

/** @type {import('eslint').Linter.Config[]} */
export default [
    includeIgnoreFile(gitignorePath),
    js.configs.recommended,
    ...svelte.configs.recommended,
    {
		languageOptions: {
			globals: { ...globals.browser, ...globals.node }
		}
	},
    {
        files: ['**/*.ts'],
        languageOptions: { parser: tseslint.parser },
        plugins: { '@typescript-eslint': tseslint.plugin },
        rules: { 'no-undef': 'off', 'no-unused-vars': 'off', '@typescript-eslint/no-unused-vars': 'warn' }
    },
    {
		files: ['**/*.svelte', '**/*.svelte.js'],
		languageOptions: { parserOptions: { svelteConfig, parser: tseslint.parser } }
	},
    {
		rules: {
			'no-unused-vars': 'warn',
			'svelte/no-navigation-without-resolve': 'off'
		}
	},
    { files: ['**/*.ts'], rules: { 'no-unused-vars': 'off' } },
    { files: ['**/*.svelte'], plugins: { '@typescript-eslint': tseslint.plugin }, rules: { 'no-unused-vars': 'off', '@typescript-eslint/no-unused-vars': 'warn' } },
    ...storybook.configs["flat/recommended"]
];
