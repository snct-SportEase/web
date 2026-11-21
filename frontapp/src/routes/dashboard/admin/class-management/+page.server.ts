import * as env from '$app/env/private';
import type { PageServerLoad } from './$types';
type ClassInfo = { id: number | string; name: string; student_count?: number; attend_count?: number };
type ClassMember = { id: string; email: string; display_name: string | null; class_id: number | null };
type EventSport = { sport_id: number; sport_name: string; template_key?: string | null; location: string };
type Sport = { id: number; name: string; templateKey: string | null; gameType: string | null };
const BACKEND_URL = env.BACKEND_URL;

export const load: PageServerLoad = async ({ fetch, locals, request }) => {
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

	const headers: Record<string, string> = {
		cookie: request.headers.get('cookie') ?? ''
	};

	const authHeader = request.headers.get('authorization');
	if (authHeader) {
		headers.authorization = authHeader;
	}

	let activeEventId: number | null = null;
	let classes: ClassInfo[] = [];
	let classMembers: ClassMember[] = [];
	let eventSports: EventSport[] = [];
	let allSports: Sport[] = [];
	let availableSports: Sport[] = [];
	let selectedClassId: number | string | null = null;
	let noonSessionName: string | null = null;
	let noonSessionSportMatched = false;
	let error: string | null = null;

	try {
        let eventData = locals.activeEvent;
        if (eventData === undefined) {
            const response = await fetch(`${BACKEND_URL}/api/events/active`, { headers });
            if (!response.ok && response.status !== 404) throw new Error('Failed to fetch active event');
            eventData = response.ok ? await response.json() : null;
        }
        activeEventId = eventData?.event_id ?? null;

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
			const [sportRes, boardGameRunsRes] = await Promise.all([
				fetch(`${BACKEND_URL}/api/events/${activeEventId}/sports`, { headers }),
				fetch(`${BACKEND_URL}/api/admin/events/${activeEventId}/board-game-runs`, { headers })
			]);
			const boardGameTypeBySportId = new Map<number, string>();
			if (boardGameRunsRes.ok) {
				const boardGameRuns: { sport_id: number; game_type: string }[] = await boardGameRunsRes.json();
				for (const run of Array.isArray(boardGameRuns) ? boardGameRuns : []) {
					boardGameTypeBySportId.set(Number(run.sport_id), run.game_type);
				}
			}
			if (sportRes.ok) {
				const sportPayload = await sportRes.json();
				eventSports = Array.isArray(sportPayload) ? sportPayload : [];
				allSports = eventSports.map((eventSport) => ({
					id: eventSport.sport_id,
					name: eventSport.sport_name,
					templateKey: eventSport.template_key ?? null,
					gameType: boardGameTypeBySportId.get(Number(eventSport.sport_id)) ?? null
				}));
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
	const availableSportMap = new Map<number, Sport>();

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
};
