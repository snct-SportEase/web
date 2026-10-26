import { defineEnvVars } from '@sveltejs/kit/env';

const optionalString = (value) => value;

export const variables = defineEnvVars({
	BACKEND_URL: { schema: optionalString },
	PUBLIC_BACKEND_URL: { public: true, schema: optionalString },
	PUBLIC_WEBPUSH_KEY: { public: true, schema: optionalString },
	PUBLIC_WEBPUSH_PUBLIC_KEY: { public: true, schema: optionalString }
});
