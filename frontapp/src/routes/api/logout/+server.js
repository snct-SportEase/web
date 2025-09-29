import { json } from '@sveltejs/kit';

/** @type {import('./$types').RequestHandler} */
export async function POST({ locals: { supabase } }) {
  const { error } = await supabase.auth.signOut();

  if (error) {
    return json({ error: 'Logout failed' }, { status: 500 });
  }

  return json({ success: true }, { status: 200 });
}
