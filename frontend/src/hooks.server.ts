import type { Handle } from '@sveltejs/kit';

const backendUrl = process.env.BACKEND_URL || 'http://127.0.0.1:8080';

export const handle: Handle = async ({ event, resolve }) => {
	if (event.url.pathname.startsWith('/api')) {
		const targetUrl = new URL(event.url.pathname + event.url.search, backendUrl);
		
		const reqHeaders = new Headers(event.request.headers);
		// Strip host to prevent backend host mismatch
		reqHeaders.delete('host');

		const init: RequestInit & { duplex?: string } = {
			method: event.request.method,
			headers: reqHeaders,
			body: event.request.method !== 'GET' && event.request.method !== 'HEAD' ? event.request.body : undefined,
			duplex: 'half'
		};

		return fetch(targetUrl.toString(), init);
	}

	return resolve(event);
};
