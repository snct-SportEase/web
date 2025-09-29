import { redirect, error } from '@sveltejs/kit';
import { PUBLIC_SUPABASE_URL, PUBLIC_SUPABASE_ANON_KEY } from '$env/static/public';
import * as pkg from '@supabase/ssr'  // ← これが一番安定する
const { createServerClient } = pkg    // SSR用は createServerClient を使う
/** @type {import('@sveltejs/kit').Handle} */
export const handle = async ({ event, resolve }) => {
  event.locals.supabase = createServerClient(PUBLIC_SUPABASE_URL, PUBLIC_SUPABASE_ANON_KEY, {
    cookies: {
      get: (key) => event.cookies.get(key),
      set: (key, value, options) => {
        event.cookies.set(key, value, { path: '/', ...options });
      },
      remove: (key, options) => {
        event.cookies.delete(key, { path: '/', ...options });
      }
    }
  });

  /**
   * a little helper that is written for convenience so that instead
   * of calling `const { data: { session } } = await event.locals.supabase.auth.getSession()`
   * you just call this `await event.locals.getSession()`
   */
  event.locals.getUser = async () => {
    const {
      data: { user },
    } = await event.locals.supabase.auth.getUser();
    return user;
  };

  const user = await event.locals.getUser();
  const { url } = event;

  // ログイン済みのユーザーがルートにアクセスした場合、ダッシュボードにリダイレクト
  if (user && url.pathname === '/') {
    throw redirect(303, '/dashboard');
  }

  // 保護されたルートへのアクセス制御
  if (!user && url.pathname.startsWith('/dashboard')) {
    throw redirect(303, '/');
  }

  return resolve(event, {
    filterSerializedResponseHeaders(name) {
      return name === 'content-range' || name === 'x-supabase-api-version';
    }
  });
};