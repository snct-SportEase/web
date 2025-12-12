import { BACKEND_URL } from '$env/static/private';

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, request, locals }) {
	const user = locals.user;
	if (!user) {
		return {
			classes: [],
			managedClass: null,
			isRoot: false
		};
	}

	const isRoot = user?.roles?.some(role => role.name === 'root');
	const classRepRole = user?.roles?.find(
		(role) => typeof role?.name === 'string' && role.name.endsWith('_rep')
	);

	const headers = {
		cookie: request.headers.get('cookie')
	};
	const authHeader = request.headers.get('Authorization');
	if (authHeader) {
		headers.Authorization = authHeader;
	}

	let classes = [];
	let managedClass = null;

	try {
		if (isRoot) {
			// Rootは全クラスを取得
			const response = await fetch(`${BACKEND_URL}/api/classes`, {
				headers
			});
			if (response.ok) {
				classes = await response.json();
			}
		} else if (classRepRole) {
			// Adminでクラス名_repロールを持っている場合は管理クラスのみ
			const managedClassResponse = await fetch(`${BACKEND_URL}/api/admin/class-team/managed-class`, {
				headers
			});
			if (managedClassResponse.ok) {
				const managedClasses = await managedClassResponse.json();
				if (Array.isArray(managedClasses) && managedClasses.length > 0) {
					classes = managedClasses;
					managedClass = managedClasses[0]; // 通常は1つのはず
				}
			}
		} else {
			// Adminでクラス名_repロールを持っていない場合は全クラス
			const response = await fetch(`${BACKEND_URL}/api/classes`, {
				headers
			});
			if (response.ok) {
				classes = await response.json();
			}
		}
	} catch (error) {
		console.error('Error loading classes:', error);
	}

	return {
		classes,
		managedClass,
		isRoot
	};
}

