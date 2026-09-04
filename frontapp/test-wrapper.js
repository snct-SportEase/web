#!/usr/bin/env node

/**
 * Test wrapper script that suppresses Vite SSR module runner transport errors
 * while preserving the Vitest exit status
 */

import { spawn } from 'child_process';
import { argv, execPath } from 'process';

const args = argv.slice(2);
const child = spawn(execPath, ['./node_modules/vitest/vitest.mjs', ...args], {
	stdio: 'pipe'
});

child.stdout.on('data', (data) => {
	const output = data.toString();
	// Always show stdout
	process.stdout.write(output);
});

child.stderr.on('data', (data) => {
	const output = data.toString();
	// Filter out transport disconnection errors that occur during cleanup
	if (!output.includes('transport was disconnected') && !output.includes('cannot call "fetchModule"')) {
		process.stderr.write(output);
	}
});

child.on('close', (code) => {
	process.exit(code);
});
