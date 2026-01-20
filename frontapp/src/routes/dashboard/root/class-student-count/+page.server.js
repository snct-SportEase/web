import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch }) {
  try {
    const response = await fetch(BACKEND_URL + `/api/classes`); // バックエンドAPIのエンドポイント
    if (!response.ok) {
      throw new Error(`Failed to fetch classes: ${response.statusText}`);
    }
    const classes = await response.json();
    return {
      classes,
    };
  } catch (error) {
    console.error('Error loading classes:', error);
    return {
      classes: [],
      error: 'クラスの読み込みに失敗しました。',
    };
  }
}
