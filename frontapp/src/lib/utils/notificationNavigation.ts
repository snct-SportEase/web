interface NotificationClient {
	navigate?(url: string): Promise<unknown>;
	focus?(): Promise<unknown>;
}
interface NotificationClients {
	matchAll(options: { type: 'window'; includeUncontrolled: boolean }): Promise<readonly NotificationClient[]>;
	openWindow?(url: string): Promise<unknown>;
}

export async function openNotificationTarget(clientsApi: NotificationClients, targetUrl: string): Promise<unknown> {
	const clientList = await clientsApi.matchAll({ type: 'window', includeUncontrolled: true });
	for (const client of clientList) {
		if (client.navigate) {
			await client.navigate(targetUrl);
		}
		if (client.focus) {
			return client.focus();
		}
	}
	if (clientsApi.openWindow) {
		return clientsApi.openWindow(targetUrl);
	}
}
