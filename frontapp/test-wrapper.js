#!/usr/bin/env node

/**
 * Test wrapper script that suppresses Vite SSR module runner transport errors
 * These errors occur after tests complete during cleanup and don't affect test results
 */

import { spawn } from 'child_process';
import { argv } from 'process';

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
