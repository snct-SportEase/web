import { BACKEND_URL } from '$env/static/private';

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, request, locals }) {
	const user = locals.user;
	const classRole = user?.roles?.find(
		(role) => typeof role?.name === 'string' && role.name.endsWith('_rep')
	);

	if (!classRole) {
		return {
			isClassRep: false,
			className: null,
			classInfo: null
		};
	}

	const className = classRole.name.slice(0, -4);

	const headers = {
		cookie: request.headers.get('cookie')
	};
	const authHeader = request.headers.get('Authorization');
	if (authHeader) {
		headers.Authorization = authHeader;
	}

	try {
		const response = await fetch(`${BACKEND_URL}/api/student/class-progress`, {
			headers
		});

		if (!response.ok) {
			if (response.status === 403) {
				return {
					isClassRep: false,
					className,
					classInfo: null,
					progress: []
				};
			}
			const errorText = await response.text();
			throw new Error(`Failed to fetch class progress: ${response.status} ${errorText}`);
		}

		const payload = await response.json();
		return {
			isClassRep: true,
			className: payload.class_name ?? className,
			classInfo: payload.class_info ?? null,
			progress: payload.progress ?? []
		};
	} catch (error) {
		console.error('Error loading class progress:', error);
		return {
			isClassRep: true,
			className,
			classInfo: null,
			progress: [],
			error: error.message
		};
	}
}

