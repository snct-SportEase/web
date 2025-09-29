import { createBrowserClient } from '@supabase/ssr'
import { PUBLIC_SUPABASE_URL, PUBLIC_SUPABASE_ANON_KEY } from '$env/static/public'

export const load = async ({ data, depends }) => {
  depends('supabase:auth')

  const supabase = createBrowserClient(
    PUBLIC_SUPABASE_URL,
    PUBLIC_SUPABASE_ANON_KEY
  )

  const {
    data: { user },
  } = await supabase.auth.getUser()

  return { supabase, user: data.user ?? user }
}
