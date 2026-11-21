import * as env from '$app/env/private';
import type { PageServerLoad } from './$types';
import type { NotificationRequest } from '$lib/types/notificationRequest';
const BACKEND_URL = env.BACKEND_URL;

export const load: PageServerLoad = async ({ fetch, request, url }) => {
  const cookie = request.headers.get('cookie') ?? '';
  const headers = {
    cookie
  };

  const returnData: { requests: NotificationRequest[]; activeRequest: NotificationRequest | null; error: string | null } = {
    requests: [],
    activeRequest: null,
    error: null
  };

  try {
    const res = await fetch(`${BACKEND_URL}/api/root/notification-requests`, { headers });
    if (res.ok) {
      const { requests } = await res.json();
      returnData.requests = requests ?? [];
    } else {
      returnData.error = '申請一覧の取得に失敗しました';
    }
  } catch (error) {
    console.error('Failed to load root notification requests:', error);
    returnData.error = '申請一覧の取得中にエラーが発生しました';
  }

  const requestedId = url.searchParams.get('request_id');
  let targetId = requestedId ? Number(requestedId) : null;

  if (!targetId && returnData.requests.length > 0) {
    targetId = returnData.requests[0].id;
  }

  if (targetId) {
    try {
      const detailRes = await fetch(`${BACKEND_URL}/api/root/notification-requests/${targetId}`, { headers });
      if (detailRes.ok) {
        const { request } = await detailRes.json();
        returnData.activeRequest = request ?? null;
      }
    } catch (error) {
      console.error('Failed to load notification request detail:', error);
    }
  }

  return returnData;
};

