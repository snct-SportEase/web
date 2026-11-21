export interface Notification {
	id: number;
	title: string;
	body: string;
	type: string;
	created_at: string;
	target_roles?: string[];
}
