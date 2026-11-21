declare module '$app/env/public' {
	export const PUBLIC_BACKEND_URL: string | undefined;
	export const PUBLIC_WEBPUSH_KEY: string | undefined;
	export const PUBLIC_WEBPUSH_PUBLIC_KEY: string | undefined;
}

declare module '$app/env/private' {
	export const BACKEND_URL: string | undefined;
}
