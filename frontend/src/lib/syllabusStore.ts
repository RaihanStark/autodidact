import { writable } from 'svelte/store';
import { api, type Syllabus } from './api';

export const syllabiList = writable<Syllabus[]>([]);
export const activeSyllabus = writable<Syllabus | null>(null);
export const activeStepId = writable<string | null>(null);
export const sidebarOpen = writable<boolean>(true);

export async function refreshSyllabi(): Promise<Syllabus[]> {
	try {
		const list = await api.listSyllabi();
		syllabiList.set(list);
		return list;
	} catch (err) {
		console.error('Failed to load syllabi', err);
		return [];
	}
}

export async function loadActiveSyllabus(id: string): Promise<Syllabus | null> {
	try {
		const data = await api.getSyllabus(id);
		activeSyllabus.set(data);
		return data;
	} catch (err) {
		console.error('Failed to load syllabus', err);
		return null;
	}
}
