const handleAuth = async ({ event, resolve }) => {
	const sessionToken = event.cookies.get('jwt'); // The cookie is named 'jwt' by the backend

	if (!sessionToken) {
		event.locals.user = null;
		return resolve(event);
	}

	const response = await event.fetch('http://back:8080/api/auth/me', {
		headers: {
			cookie: `jwt=${sessionToken}`
		}
	});

	if (response.ok) {
		event.locals.user = await response.json();
	} else {
		event.locals.user = null;
	}

	return resolve(event);
};

export const handle = handleAuth;
