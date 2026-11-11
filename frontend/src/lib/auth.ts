import { writable } from 'svelte/store';
import { getAuthToken, setAuthToken, clearAuthToken, api } from './api';

export const isAuthenticated = writable<boolean>(true);
export const isAuthModalOpen = writable<boolean>(false);
export const currentKey = writable<string>(getAuthToken());

export async function checkAuth(): Promise<boolean> {
	const token = getAuthToken();
	if (!token) {
		isAuthenticated.set(false);
		isAuthModalOpen.set(true);
		return false;
	}

	const valid = await api.verifyAuth(token);
	isAuthenticated.set(valid);
	if (!valid) {
		isAuthModalOpen.set(true);
	}
	return valid;
}

export async function submitKey(token: string): Promise<boolean> {
	const valid = await api.verifyAuth(token);
	if (valid) {
		setAuthToken(token);
		currentKey.set(token);
		isAuthenticated.set(true);
		isAuthModalOpen.set(false);
		return true;
	}
	return false;
}

export function lockWorkspace(): void {
	clearAuthToken();
	isAuthenticated.set(false);
	isAuthModalOpen.set(true);
}
