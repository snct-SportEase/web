// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			user: {
				id: string;
				email: string;
				display_name: string | null;
				class_id: number | null;
				is_profile_complete: boolean;
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
