import { writable, derived } from 'svelte/store';
import { api, type FocusItem } from './api';

export const focusList = writable<FocusItem[]>([]);

export const whatNextCount = derived(focusList, ($list) => {
	return $list.filter((item) => item.status !== 'completed').length;
});

export const todayFocusCount = whatNextCount;

export const activeFocusStepIds = derived(focusList, ($list) => {
	const map = new Map<string, FocusItem>();
	for (const item of $list) {
		if (item.step_id && item.status !== 'completed') {
			map.set(item.step_id, item);
		}
	}
	return map;
});

export async function toggleWhatNext(stepId: string): Promise<boolean> {
	let currentItem: FocusItem | undefined;
	const unsub = focusList.subscribe(list => {
		currentItem = list.find(i => i.step_id === stepId && i.status !== 'completed');
	});
	unsub();

	if (currentItem) {
		await api.deleteFocusItem(currentItem.id);
		await refreshFocus();
		return false;
	} else {
		await api.createFocusItem({ step_id: stepId, period: 'next' });
		await refreshFocus();
		return true;
	}
}

export async function refreshFocus(): Promise<FocusItem[]> {
	try {
		const items = await api.listFocusItems();
		focusList.set(items);
		return items;
	} catch (err) {
		console.error('Failed to load focus items', err);
		return [];
	}
}
