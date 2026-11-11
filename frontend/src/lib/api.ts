export type StepType = 'article' | 'video' | 'project' | 'quiz' | 'note';
export type StepStatus = 'pending' | 'in_progress' | 'completed' | 'skipped';

export type FocusPeriod = 'next' | 'today';
export type FocusStatus = 'todo' | 'in_progress' | 'completed';

export interface FocusItem {
	id: string;
	step_id?: string;
	custom_title?: string;
	title: string;
	resource_url: string;
	resources?: string[];
	module_title?: string;
	syllabus_id?: string;
	syllabus_title?: string;
	period: FocusPeriod;
	status: FocusStatus;
	notes: string;
	order_index: number;
	completed_at?: string;
	created_at: string;
	updated_at: string;
}

export interface BacklogStepItem {
	id: string;
	title: string;
	resource_url: string;
	resources?: string[];
	status: StepStatus;
	module_title: string;
	syllabus_id: string;
	syllabus_title: string;
	in_focus: boolean;
}

export interface Step {
	id: string;
	module_id: string;
	title: string;
	type: StepType;
	content: string;
	resource_url: string;
	resources?: string[];
	acceptance_criteria: string;
	estimated_minutes: number;
	order_index: number;
	status: StepStatus;
	notes: string;
	submission_url: string;
	completed_at?: string;
	created_at: string;
}

export interface Module {
	id: string;
	syllabus_id: string;
	title: string;
	description: string;
	order_index: number;
	created_at: string;
	steps: Step[];
}

export interface Syllabus {
	id: string;
	title: string;
	slug: string;
	description: string;
	is_template: boolean;
	created_at: string;
	updated_at: string;
	total_steps: number;
	completed_steps: number;
	modules?: Module[];
}

export interface ActiveStepResponse {
	step: Step | null;
	module_title?: string;
	module_id?: string;
	prev_step_id?: string | null;
	next_step_id?: string | null;
	step_index: number;
	total_steps: number;
}

const TOKEN_KEY = 'autodidact_auth_token';
const DEFAULT_TOKEN = 'autodidact-local-key';

export function getAuthToken(): string {
	if (typeof window === 'undefined') return DEFAULT_TOKEN;
	return localStorage.getItem(TOKEN_KEY) || DEFAULT_TOKEN;
}

export function setAuthToken(token: string): void {
	if (typeof window !== 'undefined') {
		localStorage.setItem(TOKEN_KEY, token);
	}
}

export function clearAuthToken(): void {
	if (typeof window !== 'undefined') {
		localStorage.removeItem(TOKEN_KEY);
	}
}

export async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
	const token = getAuthToken();
	const headers = new Headers(options.headers || {});
	
	if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
		headers.set('Content-Type', 'application/json');
	}
	if (token) {
		headers.set('Authorization', `Bearer ${token}`);
		headers.set('X-Auth-Token', token);
	}

	const response = await fetch(endpoint, {
		...options,
		headers
	});

	if (response.status === 401) {
		const errorData = await response.json().catch(() => ({}));
		throw new Error(errorData.error || 'Unauthorized: Please check your security key.');
	}

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		throw new Error(errorData.error || `HTTP error ${response.status}`);
	}

	return response.json();
}

export const api = {
	async verifyAuth(token: string): Promise<boolean> {
		try {
			const res = await fetch('/api/auth/verify', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token })
			});
			return res.ok;
		} catch {
			return false;
		}
	},

	async listSyllabi(): Promise<Syllabus[]> {
		return apiRequest<Syllabus[]>('/api/syllabi');
	},

	async getSyllabus(id: string): Promise<Syllabus> {
		return apiRequest<Syllabus>(`/api/syllabi/${id}`);
	},

	async createSyllabus(data: { title: string; description: string }): Promise<Syllabus> {
		return apiRequest<Syllabus>('/api/syllabi', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	},

	async updateSyllabus(id: string, data: { title: string; description: string }): Promise<Syllabus> {
		return apiRequest<Syllabus>(`/api/syllabi/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	},

	async deleteSyllabus(id: string): Promise<void> {
		await apiRequest(`/api/syllabi/${id}`, { method: 'DELETE' });
	},

	async duplicateSyllabus(id: string): Promise<Syllabus> {
		return apiRequest<Syllabus>(`/api/syllabi/${id}/duplicate`, { method: 'POST' });
	},

	async getActiveStep(syllabusId: string): Promise<ActiveStepResponse> {
		return apiRequest<ActiveStepResponse>(`/api/syllabi/${syllabusId}/active-step`);
	},

	async createModule(syllabusId: string, data: { title: string; description: string }): Promise<Module> {
		return apiRequest<Module>(`/api/syllabi/${syllabusId}/modules`, {
			method: 'POST',
			body: JSON.stringify(data)
		});
	},

	async updateModule(id: string, data: { title: string; description: string; order_index: number }): Promise<void> {
		await apiRequest(`/api/modules/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	},

	async deleteModule(id: string): Promise<void> {
		await apiRequest(`/api/modules/${id}`, { method: 'DELETE' });
	},

	async createStep(moduleId: string, data: {
		title: string;
		type: StepType;
		content: string;
		resource_url: string;
		resources?: string[];
		acceptance_criteria: string;
		estimated_minutes: number;
	}): Promise<Step> {
		return apiRequest<Step>(`/api/modules/${moduleId}/steps`, {
			method: 'POST',
			body: JSON.stringify(data)
		});
	},

	async getStep(id: string): Promise<Step> {
		return apiRequest<Step>(`/api/steps/${id}`);
	},

	async updateStep(id: string, data: {
		title: string;
		type: StepType;
		content: string;
		resource_url: string;
		resources?: string[];
		acceptance_criteria: string;
		estimated_minutes: number;
		order_index: number;
	}): Promise<void> {
		await apiRequest(`/api/steps/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	},

	async deleteStep(id: string): Promise<void> {
		await apiRequest(`/api/steps/${id}`, { method: 'DELETE' });
	},

	async updateStepProgress(id: string, data: {
		status: StepStatus;
		notes: string;
		submission_url: string;
	}): Promise<void> {
		await apiRequest(`/api/steps/${id}/progress`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	},

	async listFocusItems(period?: string): Promise<FocusItem[]> {
		const q = period ? `?period=${encodeURIComponent(period)}` : '';
		return apiRequest<FocusItem[]>(`/api/focus${q}`);
	},

	async createFocusItem(data: {
		step_id?: string;
		custom_title?: string;
		resource_url?: string;
		resources?: string[];
		period?: FocusPeriod;
		notes?: string;
	}): Promise<FocusItem> {
		return apiRequest<FocusItem>('/api/focus', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	},

	async updateFocusItem(id: string, data: {
		status?: FocusStatus;
		period?: FocusPeriod;
		notes?: string;
		resources?: string[];
	}): Promise<FocusItem> {
		return apiRequest<FocusItem>(`/api/focus/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	},

	async deleteFocusItem(id: string): Promise<void> {
		await apiRequest(`/api/focus/${id}`, { method: 'DELETE' });
	},

	async reorderFocusItems(ids: string[]): Promise<void> {
		await apiRequest('/api/focus/reorder', {
			method: 'PUT',
			body: JSON.stringify({ ids })
		});
	},

	async listBacklogItems(): Promise<BacklogStepItem[]> {
		return apiRequest<BacklogStepItem[]>('/api/focus/backlog-items');
	}
};
