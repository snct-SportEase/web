export async function openNotificationTarget(clientsApi, targetUrl) {
	const clientList = await clientsApi.matchAll({ type: 'window', includeUncontrolled: true });
	for (const client of clientList) {
		if ('navigate' in client) {
			await client.navigate(targetUrl);
		}
		if ('focus' in client) {
			return client.focus();
		}
	}
	if (clientsApi.openWindow) {
		return clientsApi.openWindow(targetUrl);
	}
}
