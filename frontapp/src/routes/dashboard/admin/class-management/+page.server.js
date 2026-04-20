import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, locals, request }) {
	const user = locals.user;

	if (!user) {
		return {
			user: null,
			isAdmin: false,
			classes: [],
			classMembers: [],
			eventSports: [],
			allSports: [],
			activeEventId: null,
			selectedClassId: null
		};
	}

	const isAdmin =
		user.roles?.some((role) => role.name === 'admin' || role.name === 'root') || false;

	const headers = {
		cookie: request.headers.get('cookie') ?? ''
	};

	const authHeader = request.headers.get('authorization');
	if (authHeader) {
		headers.authorization = authHeader;
	}

	let activeEventId = null;
	let classes = [];
	let classMembers = [];
	let eventSports = [];
	let allSports = [];
	let availableSports = [];
	let selectedClassId = null;
	let noonSessionName = null;
	let noonSessionSportMatched = false;
	let error = null;

	try {
		const activeEventRes = await fetch(`${BACKEND_URL}/api/events/active`, {
			headers
		});
		if (activeEventRes.ok) {
			const eventData = await activeEventRes.json();
			activeEventId = eventData?.event_id ?? null;
		} else if (activeEventRes.status !== 404) {
			const text = await activeEventRes.text();
			throw new Error(`Failed to fetch active event: ${activeEventRes.status} ${text}`);
		}

		const classesRes = await fetch(`${BACKEND_URL}/api/admin/class-team/managed-class`, {
			headers
		});

		if (!classesRes.ok) {
			const text = await classesRes.text();
			throw new Error(`Failed to fetch managed classes: ${classesRes.status} ${text}`);
		}

		const classesPayload = await classesRes.json();
		classes = Array.isArray(classesPayload) ? classesPayload : [];
		if (classes.length > 0) {
			selectedClassId = classes[0]?.id ?? null;

			if (selectedClassId != null) {
				const membersRes = await fetch(
					`${BACKEND_URL}/api/admin/class-team/classes/${selectedClassId}/members`,
					{ headers }
				);
				if (membersRes.ok) {
					const membersPayload = await membersRes.json();
					classMembers = Array.isArray(membersPayload) ? membersPayload : [];
				}
			}
		}

		if (activeEventId) {
			const sportRes = await fetch(`${BACKEND_URL}/api/events/${activeEventId}/sports`, {
				headers
			});
			if (sportRes.ok) {
				const sportPayload = await sportRes.json();
				eventSports = Array.isArray(sportPayload) ? sportPayload : [];
			}

			const allSportRes = await fetch(`${BACKEND_URL}/api/admin/allsports`, {
				headers
			});
			if (allSportRes.ok) {
				const allSportPayload = await allSportRes.json();
				allSports = Array.isArray(allSportPayload) ? allSportPayload : [];
			}

			const noonSessionRes = await fetch(
				`${BACKEND_URL}/api/student/events/${activeEventId}/noon-game/session`,
				{ headers }
			);
			if (noonSessionRes.ok) {
				const noonSessionPayload = await noonSessionRes.json();
				noonSessionName = noonSessionPayload?.session?.name?.trim() || null;
			}
		}
	} catch (err) {
		console.error('Failed to load class management data:', err);
		error = err instanceof Error ? err.message : 'データの取得に失敗しました';
	}

	classes = Array.isArray(classes) ? classes : [];
	classMembers = Array.isArray(classMembers) ? classMembers : [];
	eventSports = Array.isArray(eventSports) ? eventSports : [];
	allSports = Array.isArray(allSports) ? allSports : [];
	availableSports = [];

	const assignedSportIds = new Set(
		eventSports
			.map((eventSport) => Number(eventSport?.sport_id))
			.filter((sportId) => !Number.isNaN(sportId))
	);
	const availableSportMap = new Map();

	for (const sport of allSports) {
		if (assignedSportIds.has(sport.id)) {
			availableSportMap.set(sport.id, sport);
		}
	}

	if (noonSessionName) {
		const matchedNoonSport = allSports.find((sport) => sport.name === noonSessionName);
		if (matchedNoonSport) {
			availableSportMap.set(matchedNoonSport.id, matchedNoonSport);
			noonSessionSportMatched = true;
		}
	}

	availableSports = Array.from(availableSportMap.values());

	if (selectedClassId != null && typeof selectedClassId !== 'number') {
		const parsed = Number(selectedClassId);
		selectedClassId = Number.isNaN(parsed) ? null : parsed;
	}

	return {
		user,
		isAdmin,
		activeEventId,
		classes,
		classMembers,
		eventSports,
		allSports,
		availableSports,
		selectedClassId,
		noonSessionName,
		noonSessionSportMatched,
		error
	};
}
