#!/usr/bin/env node

/**
 * Test wrapper script that suppresses Vite SSR module runner transport errors
 * and ensures Playwright browsers are installed before running tests
 */

import { spawn, spawnSync } from 'child_process';
import { argv } from 'process';

// Check if we're in a CI environment
const isCI = process.env.CI === 'true';

// In CI, ensure Playwright browsers are installed
if (isCI) {
	console.log('Installing Playwright browsers...');
	const installResult = spawnSync('npx', ['playwright', 'install', '--with-deps'], {
		stdio: 'inherit'
	});
	
	if (installResult.status !== 0) {
		console.error('Failed to install Playwright browsers');
		process.exit(1);
	}
}

const args = argv.slice(2);
const child = spawn('npm', ['run', 'test:unit', '--', ...args], {
	stdio: 'pipe'
});

let stdoutBuffer = '';
let stderrBuffer = '';

child.stdout.on('data', (data) => {
	const output = data.toString();
	stdoutBuffer += output;
	// Always show stdout
	process.stdout.write(output);
});

child.stderr.on('data', (data) => {
	const output = data.toString();
	// Filter out transport disconnection errors that occur during cleanup
	if (!output.includes('transport was disconnected') && !output.includes('cannot call "fetchModule"')) {
		process.stderr.write(output);
	}
	stderrBuffer += output;
});

child.on('close', (code) => {
	process.exit(code);
});
