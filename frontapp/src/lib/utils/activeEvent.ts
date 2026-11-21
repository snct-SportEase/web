export interface ActiveEvent {
    id: number;
    name: string;
    year?: number;
    season?: string;
    status?: string;
    is_rainy_mode?: boolean;
    hide_scores?: boolean;
    test_run_state?: string;
    competition_guidelines_pdf_url?: string | null;
    survey_url?: string | null;
    is_survey_published?: boolean;
    is_mic_voting_enabled?: boolean;
    start_date?: string;
    end_date?: string;
    duplicate_registration_threshold?: number;
}

export type ActiveEventResponse = Omit<ActiveEvent, 'id' | 'name'> & {
    event_id: number | null;
    event_name: string | null;
    id?: number;
};

export function normalizeActiveEvent(payload: ActiveEventResponse | null): ActiveEvent | null {
    if (!payload?.event_id) return null;
    return { ...payload, id: payload.event_id, name: payload.event_name ?? '' };
}
