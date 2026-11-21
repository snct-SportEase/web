// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			activeEvent: import('$lib/utils/activeEvent.js').ActiveEventResponse | null;
			user: {
				id: string;
				email: string;
				notification_filters?: string[];
				display_name: string | null;
				class_id: number | null;
				is_profile_complete: boolean;
				is_init_root_first_login?: boolean;
				roles?: Array<{
					id: number;
					name: string;
				}>;
				created_at: string;
				updated_at: string;
			} | null;
		}
		// interface PageData {}
		// interface Platform {}
	}
}

export {};
