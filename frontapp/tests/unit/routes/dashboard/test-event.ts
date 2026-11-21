import type { RequestEvent } from '@sveltejs/kit';

export function makeEvent(token: string, path = '/dashboard/student/my-page'): RequestEvent {
 const url = new URL(`http://localhost${path}`);
 return {
  url, request: new Request(url), fetch, getClientAddress: () => '127.0.0.1',
  cookies: {
   get: () => token, getAll: () => [], set: () => {}, delete: () => {},
   parse: () => { throw new Error('予期しない Cookie 解析'); },
   serialize: () => { throw new Error('予期しない Cookie 生成'); }
  },
  locals: { user: null, activeEvent: null }, params: {}, platform: undefined,
  route: { id: null }, setHeaders: () => {}, isDataRequest: false, isSubRequest: false,
  isRemoteRequest: false,
  get tracing(): never { throw new Error('予期しないトレーシング'); }
 };
}
