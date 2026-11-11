<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api, type Syllabus, type Module, type Step, type StepStatus } from '#lib/api';
	import { activeSyllabus, refreshSyllabi, loadActiveSyllabus } from '#lib/syllabusStore';
	import { refreshFocus, activeFocusStepIds, toggleWhatNext } from '#lib/focusStore';
	import Markdown from '#lib/components/Markdown.svelte';
	import { 
		Play, 
		Plus, 
		Edit3, 
		Trash2, 
		CheckCircle2, 
		Circle, 
		ExternalLink, 
		FastForward, 
		Save, 
		Check, 
		ArrowLeft, 
		ArrowRight, 
		Copy,
		Link as LinkIcon,
		StickyNote,
		CornerDownLeft,
		Target,
		Calendar,
		X
	} from '@lucide/svelte';

	let syllabusId = $derived(page.params.id);
	let stepQueryParam = $derived(page.url.searchParams.get('step'));

	let syllabus = $state<Syllabus | null>(null);
	let loading = $state(true);
	let error = $state('');

	// Active step state
	let currentStep = $state<Step | null>(null);
	let currentModule = $state<Module | null>(null);
	let notesText = $state('');
	let submissionUrl = $state('');
	let savingNotes = $state(false);
	let notesSavedNotice = $state(false);
	let previewNotes = $state(false);

	// Flattened step list for easy sequential next/prev
	let flatSteps = $state<{ step: Step; moduleTitle: string }[]>([]);
	let currentFlatIndex = $state(-1);

	// Quick add state (per module)
	let quickAddInputs = $state<Record<string, string>>({});
	let quickAddLoading = $state<Record<string, boolean>>({});
	let quickAddInputRefs: Record<string, HTMLInputElement> = $state({});

	// Section modal state
	let showModuleModal = $state(false);
	let editingModule = $state<Module | null>(null);
	let moduleTitle = $state('');
	let isSavingModule = $state(false);

	// Item modal state
	let showStepModal = $state(false);
	let editingStep = $state<Step | null>(null);
	let targetModuleId = $state('');
	let stepTitle = $state('');
	let stepContent = $state('');
	let stepResources = $state<string[]>(['']);
	let isSavingStep = $state(false);

	// Inline resource state in workspace
	let newResourceInput = $state('');
	let isAddingResource = $state(false);

	// Syllabus edit modal state
	let showSyllabusModal = $state(false);
	let sylTitle = $state('');
	let sylDesc = $state('');

	onMount(() => {
		loadData();
	});

	$effect(() => {
		if (syllabusId) {
			loadData();
		}
	});

	$effect(() => {
		if (syllabus && flatSteps.length > 0) {
			updateActiveStepFromQuery();
		}
	});

	async function loadData(showLoading = true) {
		try {
			if (showLoading) loading = true;
			error = '';
			const data = await api.getSyllabus(syllabusId);
			syllabus = data;
			activeSyllabus.set(data);
			if (data) {
				sylTitle = data.title;
				sylDesc = data.description;

				const flattened: { step: Step; moduleTitle: string }[] = [];
				for (const mod of data.modules || []) {
					for (const st of mod.steps || []) {
						flattened.push({ step: st, moduleTitle: mod.title });
					}
				}
				flatSteps = flattened;
				updateActiveStepFromQuery();
			}
		} catch (err: any) {
			error = err.message || 'Failed to load queue';
		} finally {
			if (showLoading) loading = false;
		}
	}

	function updateActiveStepFromQuery() {
		if (!stepQueryParam) {
			currentStep = null;
			currentModule = null;
			currentFlatIndex = -1;
			return;
		}

		const idx = flatSteps.findIndex(f => f.step.id === stepQueryParam);
		if (idx !== -1) {
			currentFlatIndex = idx;
			currentStep = flatSteps[idx].step;
			notesText = currentStep.notes || '';
			submissionUrl = currentStep.submission_url || '';
			currentModule = syllabus?.modules?.find(m => m.id === currentStep?.module_id) || null;
		} else {
			currentStep = null;
			currentModule = null;
			currentFlatIndex = -1;
		}
	}

	function navigateToStep(stepId: string) {
		goto(`/syllabus/${syllabusId}?step=${stepId}`);
	}

	async function handleQuickAdd(moduleId: string) {
		const inputVal = quickAddInputs[moduleId]?.trim();
		if (!inputVal) return;

		quickAddLoading[moduleId] = true;
		try {
			const urlMatches = inputVal.match(/https?:\/\/[^\s]+/g);
			let resources: string[] = [];
			let resourceUrl = '';
			let cleanTitle = inputVal;

			if (urlMatches && urlMatches.length > 0) {
				resources = urlMatches;
				resourceUrl = resources[0];

				if (inputVal.startsWith('http://') || inputVal.startsWith('https://')) {
					try {
						const urlObj = new URL(resources[0]);
						const rawPath = decodeURIComponent(urlObj.pathname.slice(1)).replace(/[-_]+/g, ' ').trim();
						cleanTitle = rawPath ? (urlObj.hostname + ' ' + rawPath).slice(0, 50) : urlObj.hostname;
					} catch {
						cleanTitle = resources[0];
					}
				}
			}

			// Remove any leading or trailing dashes / strip characters and multiple dashes
			cleanTitle = cleanTitle
				.replace(/^[\s\-–—*•]+/, '')
				.replace(/[\s\-–—*•]+$/, '')
				.replace(/[\-–—]{2,}/g, ' ')
				.replace(/\s+/g, ' ')
				.trim();

			await api.createStep(moduleId, {
				title: cleanTitle || 'Untitled Item',
				type: 'article',
				content: '',
				resource_url: resourceUrl,
				resources: resources,
				acceptance_criteria: '',
				estimated_minutes: 0
			});

			quickAddInputs[moduleId] = '';
			await loadData(false);
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to add item');
		} finally {
			quickAddLoading[moduleId] = false;
			await tick();
			quickAddInputRefs[moduleId]?.focus();
		}
	}

	async function updateStatus(step: Step, newStatus: StepStatus) {
		try {
			await api.updateStepProgress(step.id, {
				status: newStatus,
				notes: notesText || step.notes || '',
				submission_url: submissionUrl || step.submission_url || ''
			});
			step.status = newStatus;
			if (currentStep && currentStep.id === step.id) {
				currentStep.status = newStatus;
			}
			await loadActiveSyllabus(syllabusId);
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to update status');
		}
	}

	async function saveNotes(silent = false) {
		if (!currentStep) return;
		savingNotes = true;
		try {
			await api.updateStepProgress(currentStep.id, {
				status: currentStep.status,
				notes: notesText,
				submission_url: submissionUrl
			});
			currentStep.notes = notesText;
			currentStep.submission_url = submissionUrl;
			if (!silent) {
				notesSavedNotice = true;
				setTimeout(() => { notesSavedNotice = false; }, 2000);
			}
		} catch (err: any) {
			alert(err.message || 'Failed to save notes');
		} finally {
			savingNotes = false;
		}
	}

	async function handleDoneAndAdvance() {
		if (!currentStep) return;
		await updateStatus(currentStep, 'completed');
		if (currentFlatIndex < flatSteps.length - 1) {
			navigateToStep(flatSteps[currentFlatIndex + 1].step.id);
		}
	}

	async function handleSkipAndAdvance() {
		if (!currentStep) return;
		await updateStatus(currentStep, 'skipped');
		if (currentFlatIndex < flatSteps.length - 1) {
			navigateToStep(flatSteps[currentFlatIndex + 1].step.id);
		}
	}

	async function handleToggleFocus(stepId: string) {
		try {
			await toggleWhatNext(stepId);
		} catch (err: any) {
			alert(err.message || 'Failed to update What Next');
		}
	}

	function openCreateModule() {
		editingModule = null;
		moduleTitle = '';
		showModuleModal = true;
	}

	function openEditModule(mod: Module) {
		editingModule = mod;
		moduleTitle = mod.title;
		showModuleModal = true;
	}

	async function handleSaveModule(e: Event) {
		e.preventDefault();
		if (!moduleTitle.trim()) return;
		isSavingModule = true;
		try {
			if (editingModule) {
				await api.updateModule(editingModule.id, {
					title: moduleTitle.trim(),
					description: '',
					order_index: editingModule.order_index
				});
			} else {
				await api.createModule(syllabusId, {
					title: moduleTitle.trim(),
					description: ''
				});
			}
			showModuleModal = false;
			await loadData();
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to save section');
		} finally {
			isSavingModule = false;
		}
	}

	async function handleDeleteModule(modId: string, title: string) {
		if (!confirm(`Delete section "${title}"?`)) return;
		try {
			await api.deleteModule(modId);
			await loadData();
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to delete section');
		}
	}

	function openCreateStep(moduleId: string) {
		editingStep = null;
		targetModuleId = moduleId;
		stepTitle = '';
		stepContent = '';
		stepResources = [''];
		showStepModal = true;
	}

	function openEditStep(step: Step) {
		editingStep = step;
		targetModuleId = step.module_id;
		stepTitle = step.title;
		stepContent = step.content;
		if (step.resources && step.resources.length > 0) {
			stepResources = [...step.resources];
		} else if (step.resource_url) {
			stepResources = [step.resource_url];
		} else {
			stepResources = [''];
		}
		showStepModal = true;
	}

	function addResourceLink() {
		stepResources = [...stepResources, ''];
	}

	function removeResourceLink(index: number) {
		stepResources = stepResources.filter((_, i) => i !== index);
		if (stepResources.length === 0) {
			stepResources = [''];
		}
	}

	async function handleSaveStep(e: Event) {
		e.preventDefault();
		const cleanedTitle = stepTitle
			.replace(/^[\s\-–—*•]+/, '')
			.replace(/[\s\-–—*•]+$/, '')
			.replace(/[\-–—]{2,}/g, ' ')
			.replace(/\s+/g, ' ')
			.trim();
		if (!cleanedTitle) return;
		isSavingStep = true;
		try {
			const cleanResources = stepResources
				.map(r => r.trim())
				.filter(r => r.length > 0);
			const primaryUrl = cleanResources[0] || '';

			if (editingStep) {
				await api.updateStep(editingStep.id, {
					title: cleanedTitle,
					type: 'article',
					content: stepContent.trim(),
					resource_url: primaryUrl,
					resources: cleanResources,
					acceptance_criteria: '',
					estimated_minutes: 0,
					order_index: editingStep.order_index
				});
			} else {
				await api.createStep(targetModuleId, {
					title: cleanedTitle,
					type: 'article',
					content: stepContent.trim(),
					resource_url: primaryUrl,
					resources: cleanResources,
					acceptance_criteria: '',
					estimated_minutes: 0
				});
			}
			showStepModal = false;
			await loadData();
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to save item');
		} finally {
			isSavingStep = false;
		}
	}

	async function handleAddResourceToCurrentStep() {
		if (!currentStep || !newResourceInput.trim()) return;
		const link = newResourceInput.trim();
		const existing = (currentStep.resources && currentStep.resources.length > 0)
			? [...currentStep.resources]
			: (currentStep.resource_url ? [currentStep.resource_url] : []);
		if (!existing.includes(link)) {
			existing.push(link);
		}
		isAddingResource = true;
		try {
			await api.updateStep(currentStep.id, {
				title: currentStep.title,
				type: currentStep.type,
				content: currentStep.content,
				resource_url: existing[0] || '',
				resources: existing,
				acceptance_criteria: currentStep.acceptance_criteria,
				estimated_minutes: currentStep.estimated_minutes,
				order_index: currentStep.order_index
			});
			currentStep.resources = existing;
			currentStep.resource_url = existing[0] || '';
			newResourceInput = '';
			await loadData(false);
		} catch (err: any) {
			alert(err.message || 'Failed to add link');
		} finally {
			isAddingResource = false;
		}
	}

	async function handleRemoveResourceFromCurrentStep(index: number) {
		if (!currentStep) return;
		const existing = (currentStep.resources && currentStep.resources.length > 0)
			? [...currentStep.resources]
			: (currentStep.resource_url ? [currentStep.resource_url] : []);
		const updated = existing.filter((_, i) => i !== index);
		try {
			await api.updateStep(currentStep.id, {
				title: currentStep.title,
				type: currentStep.type,
				content: currentStep.content,
				resource_url: updated[0] || '',
				resources: updated,
				acceptance_criteria: currentStep.acceptance_criteria,
				estimated_minutes: currentStep.estimated_minutes,
				order_index: currentStep.order_index
			});
			currentStep.resources = updated;
			currentStep.resource_url = updated[0] || '';
			await loadData(false);
		} catch (err: any) {
			alert(err.message || 'Failed to remove link');
		}
	}

	async function handleDeleteStep(stepId: string, title: string) {
		if (!confirm(`Delete "${title}"?`)) return;
		try {
			await api.deleteStep(stepId);
			if (currentStep?.id === stepId) {
				goto(`/syllabus/${syllabusId}`);
			}
			await loadData();
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to delete');
		}
	}

	async function handleSaveSyllabus(e: Event) {
		e.preventDefault();
		if (!sylTitle.trim()) return;
		try {
			await api.updateSyllabus(syllabusId, {
				title: sylTitle.trim(),
				description: sylDesc.trim()
			});
			showSyllabusModal = false;
			await loadData();
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to update');
		}
	}

	async function handleDuplicate() {
		try {
			const dup = await api.duplicateSyllabus(syllabusId);
			await refreshSyllabi();
			goto(`/syllabus/${dup.id}`);
		} catch (err: any) {
			alert(err.message || 'Failed to duplicate');
		}
	}

	async function handleDeleteSyllabus() {
		if (!syllabus) return;
		if (!confirm(`Delete "${syllabus.title}"?`)) return;
		try {
			await api.deleteSyllabus(syllabusId);
			const list = await refreshSyllabi();
			if (list.length > 0) {
				goto(`/syllabus/${list[0].id}`);
			} else {
				goto('/');
			}
		} catch (err: any) {
			alert(err.message || 'Failed to delete');
		}
	}
</script>

<div class="h-full w-full">
	{#if loading}
		<div class="p-8 space-y-4 max-w-4xl mx-auto animate-pulse">
			<div class="h-8 bg-zinc-800 rounded-lg w-1/3"></div>
			<div class="h-20 bg-zinc-800/60 rounded-xl"></div>
			<div class="h-64 bg-zinc-800/40 rounded-xl"></div>
		</div>
	{:else if error || !syllabus}
		<div class="p-8 max-w-xl mx-auto text-center space-y-4">
			<p class="text-sm text-red-400">{error || 'List not found'}</p>
			<a href="/" class="inline-flex rounded-lg bg-zinc-800 px-4 py-2 text-xs text-zinc-200">Return home</a>
		</div>
	{:else if currentStep}
		<!-- ITEM WORKSPACE VIEW -->
		<div class="max-w-4xl mx-auto px-6 py-8 space-y-6">
			<!-- Header & Status -->
			<div class="space-y-4 pb-6 border-b border-zinc-800/80">
				<div class="flex items-center justify-between gap-4 text-xs text-zinc-400">
					<div class="flex items-center gap-2 truncate">
						<a href="/syllabus/{syllabusId}" class="hover:text-white transition-colors">
							← Back to Overview
						</a>
						<span>•</span>
						<span class="font-medium text-zinc-300 truncate">{currentModule?.title}</span>
						<span>•</span>
						<span class="font-mono text-zinc-400">Item {currentFlatIndex + 1} of {flatSteps.length}</span>
					</div>

					<button
						type="button"
						onclick={() => currentStep && openEditStep(currentStep)}
						class="p-1 rounded text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
						title="Edit item title/link"
					>
						<Edit3 class="h-3.5 w-3.5" />
					</button>
				</div>

				<h1 class="text-2xl font-bold tracking-tight text-white leading-snug">
					{currentStep.title}
				</h1>

				<!-- Status Pills & External Link -->
				<div class="flex flex-wrap items-center justify-between gap-3 pt-1">
					<div class="flex items-center gap-1.5 p-1 rounded-xl bg-zinc-900 border border-zinc-800">
						<button
							type="button"
							onclick={() => currentStep && updateStatus(currentStep, 'pending')}
							class="px-3 py-1 text-xs rounded-lg font-medium transition-colors {currentStep.status === 'pending' ? 'bg-zinc-700 text-white shadow-xs' : 'text-zinc-400 hover:text-zinc-200'}"
						>
							To Do
						</button>
						<button
							type="button"
							onclick={() => currentStep && updateStatus(currentStep, 'in_progress')}
							class="px-3 py-1 text-xs rounded-lg font-medium transition-colors {currentStep.status === 'in_progress' ? 'bg-indigo-600 text-white shadow-xs' : 'text-zinc-400 hover:text-zinc-200'}"
						>
							Doing
						</button>
						<button
							type="button"
							onclick={() => currentStep && updateStatus(currentStep, 'completed')}
							class="px-3 py-1 text-xs rounded-lg font-medium transition-colors {currentStep.status === 'completed' ? 'bg-emerald-600 text-white shadow-xs' : 'text-zinc-400 hover:text-zinc-200'}"
						>
							Done ✓
						</button>
						<button
							type="button"
							onclick={() => currentStep && updateStatus(currentStep, 'skipped')}
							class="px-3 py-1 text-xs rounded-lg font-medium transition-colors {currentStep.status === 'skipped' ? 'bg-amber-600 text-white shadow-xs' : 'text-zinc-400 hover:text-zinc-200'}"
						>
							Skipped ↷
						</button>
					</div>

					<button
						type="button"
						onclick={() => currentStep && handleToggleFocus(currentStep.id)}
						class="inline-flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs font-medium transition-colors {$activeFocusStepIds.has(currentStep.id) ? 'border border-indigo-500/50 bg-indigo-500/20 text-indigo-300 font-semibold' : 'border border-zinc-700 bg-zinc-800 text-zinc-300 hover:text-white hover:bg-zinc-750'}"
						title={$activeFocusStepIds.has(currentStep.id) ? 'Queued in What Next (click to remove)' : 'Add to What Next'}
					>
						<Target class="h-3.5 w-3.5" />
						<span>{$activeFocusStepIds.has(currentStep.id) ? '✓ In What Next' : '+ What Next'}</span>
					</button>
				</div>

				<!-- Resources / Unlimited Multiple Links -->
				<div class="space-y-2 pt-2 border-t border-zinc-800/60">
					<div class="flex items-center justify-between">
						<span class="text-[11px] font-bold uppercase tracking-wider text-zinc-500 flex items-center gap-1.5">
							<LinkIcon class="h-3 w-3 text-zinc-500" />
							<span>Resources ({ (currentStep.resources && currentStep.resources.length > 0) ? currentStep.resources.length : (currentStep.resource_url ? 1 : 0) })</span>
						</span>
					</div>

					{#if (currentStep.resources && currentStep.resources.length > 0) || currentStep.resource_url}
						{@const allLinks = (currentStep.resources && currentStep.resources.length > 0) ? currentStep.resources : [currentStep.resource_url]}
						<div class="space-y-1.5">
							{#each allLinks as link, idx}
								<div class="group flex items-center justify-between gap-2 rounded-lg bg-zinc-900/60 border border-zinc-800/80 px-3 py-1.5 text-xs hover:border-zinc-700 transition-colors">
									<div class="flex items-center gap-2 min-w-0 flex-1">
										<span class="text-zinc-500 font-mono text-[11px] font-medium shrink-0">Link:</span>
										<a
											href={link}
											target="_blank"
											rel="noopener noreferrer"
											class="inline-flex items-center gap-1.5 font-mono text-xs text-indigo-400 hover:text-indigo-300 hover:underline break-all truncate"
										>
											<span class="truncate">{link}</span>
											<ExternalLink class="h-3 w-3 shrink-0" />
										</a>
									</div>
									<button
										type="button"
										onclick={() => handleRemoveResourceFromCurrentStep(idx)}
										class="text-zinc-600 hover:text-red-400 p-0.5 opacity-0 group-hover:opacity-100 transition-opacity"
										title="Remove link"
									>
										<X class="h-3.5 w-3.5" />
									</button>
								</div>
							{/each}
						</div>
					{/if}

					<!-- Inline Quick Add Link input -->
					<form onsubmit={(e) => { e.preventDefault(); handleAddResourceToCurrentStep(); }} class="flex items-center gap-2 pt-1">
						<span class="text-zinc-500 font-mono text-[11px] shrink-0">Link:</span>
						<input
							type="url"
							bind:value={newResourceInput}
							placeholder="https://... (press Enter to attach)"
							class="flex-1 rounded-lg border border-zinc-800 bg-zinc-900/50 px-3 py-1 font-mono text-xs text-zinc-200 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden"
						/>
						<button
							type="submit"
							disabled={!newResourceInput.trim() || isAddingResource}
							class="inline-flex items-center gap-1 rounded-lg bg-zinc-800 px-2.5 py-1 text-xs text-zinc-300 hover:bg-zinc-700 disabled:opacity-40 disabled:cursor-not-allowed"
						>
							<Plus class="h-3 w-3" />
							<span>Add</span>
						</button>
					</form>
				</div>
			</div>

			<!-- Brief / Context (if user added one) -->
			{#if currentStep.content}
				<div class="space-y-2">
					<div class="rounded-xl border border-zinc-800/80 bg-zinc-900/40 p-4 text-xs text-zinc-300 leading-relaxed">
						<Markdown content={currentStep.content} />
					</div>
				</div>
			{/if}

			<!-- Personal Scratchpad / Reading Notes -->
			<div class="space-y-2 pt-2">
				<div class="flex items-center justify-between">
					<h3 class="text-xs font-bold uppercase tracking-wider text-zinc-400 flex items-center gap-1.5">
						<StickyNote class="h-3.5 w-3.5 text-zinc-400" />
						<span>Reading Notes & Takeaways</span>
					</h3>
					<div class="flex items-center gap-2">
						<button
							type="button"
							onclick={() => previewNotes = !previewNotes}
							class="text-xs text-zinc-400 hover:text-white underline underline-offset-4"
						>
							{previewNotes ? 'Edit Raw' : 'Preview Markdown'}
						</button>
						<button
							type="button"
							onclick={() => saveNotes(false)}
							disabled={savingNotes}
							class="inline-flex items-center gap-1 rounded-lg bg-zinc-800 px-3 py-1 text-xs text-zinc-200 hover:bg-zinc-700"
						>
							<Save class="h-3 w-3" />
							<span>{savingNotes ? 'Saving...' : 'Save Notes'}</span>
						</button>
					</div>
				</div>

				{#if notesSavedNotice}
					<p class="text-xs text-emerald-400 flex items-center gap-1">
						<Check class="h-3 w-3" /> Notes saved to local database!
					</p>
				{/if}

				{#if previewNotes}
					<div class="min-h-[160px] rounded-xl border border-zinc-800 bg-zinc-900/50 p-4 text-xs leading-relaxed">
						{#if notesText.trim()}
							<Markdown content={notesText} />
						{:else}
							<span class="text-zinc-500 italic">No notes written yet. Jot down chapter summaries, takeaways, code snippets...</span>
						{/if}
					</div>
				{:else}
					<textarea
						bind:value={notesText}
						rows="8"
						placeholder="Jot down notes while reading or watching: key concepts, commands, quotes, code snippets..."
						class="w-full rounded-xl border border-zinc-800 bg-zinc-900/60 p-3.5 font-mono text-xs text-zinc-100 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden leading-relaxed"
					></textarea>
				{/if}

				<!-- Optional Reference Link -->
				<div class="pt-2">
					<label for="sub-url" class="block text-[11px] font-medium text-zinc-400 mb-1 flex items-center gap-1">
						<LinkIcon class="h-3 w-3" />
						<span>Project or PR Link (optional)</span>
					</label>
					<input
						id="sub-url"
						type="url"
						bind:value={submissionUrl}
						placeholder="https://github.com/..."
						class="w-full rounded-xl border border-zinc-800 bg-zinc-900/60 px-3 py-1.5 text-xs text-zinc-200 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden"
					/>
				</div>
			</div>

			<!-- Sequential Bottom Bar -->
			<div class="pt-6 border-t border-zinc-800 flex items-center justify-between gap-3">
				<button
					type="button"
					onclick={() => currentFlatIndex > 0 && navigateToStep(flatSteps[currentFlatIndex - 1].step.id)}
					disabled={currentFlatIndex <= 0}
					class="inline-flex items-center gap-1.5 rounded-xl border border-zinc-800 bg-zinc-900 px-3.5 py-2 text-xs font-medium text-zinc-300 hover:bg-zinc-800 disabled:opacity-30 disabled:cursor-not-allowed"
				>
					<ArrowLeft class="h-3.5 w-3.5" />
					<span>Previous</span>
				</button>

				<div class="flex items-center gap-2">
					<button
						type="button"
						onclick={handleSkipAndAdvance}
						class="px-3 py-2 text-xs font-medium text-zinc-400 hover:text-amber-300 transition-colors"
					>
						Skip
					</button>

					<button
						type="button"
						onclick={handleDoneAndAdvance}
						class="inline-flex items-center gap-1.5 rounded-xl bg-indigo-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-indigo-500"
					>
						<CheckCircle2 class="h-3.5 w-3.5" />
						<span>Mark Done & Next</span>
					</button>

					{#if currentFlatIndex < flatSteps.length - 1}
						<button
							type="button"
							onclick={() => navigateToStep(flatSteps[currentFlatIndex + 1].step.id)}
							class="p-2 rounded-xl border border-zinc-800 bg-zinc-900 text-zinc-400 hover:text-white hover:bg-zinc-800"
							title="Next item"
						>
							<ArrowRight class="h-3.5 w-3.5" />
						</button>
					{/if}
				</div>
			</div>
		</div>
	{:else}
		<!-- QUEUE LIST OVERVIEW -->
		{@const totalCount = flatSteps.length}
		{@const doneCount = flatSteps.filter(f => f.step.status === 'completed').length}
		{@const progressPct = totalCount > 0 ? Math.round((doneCount / totalCount) * 100) : 0}
		<div class="max-w-4xl mx-auto px-6 py-8 space-y-8">
			<!-- Header Box -->
			<div class="rounded-2xl border border-zinc-800 bg-zinc-900/40 p-6 space-y-4">
				<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
					<div>
						<div class="flex items-center gap-2">
							<h1 class="text-2xl font-bold text-white">{syllabus.title}</h1>
							<button
								type="button"
								onclick={() => showSyllabusModal = true}
								class="p-1 rounded text-zinc-400 hover:text-white"
								title="Edit title & notes"
							>
								<Edit3 class="h-4 w-4" />
							</button>
						</div>
						<p class="mt-1 text-xs text-zinc-400 leading-relaxed max-w-xl">{syllabus.description || 'Personal queue of things to learn next.'}</p>
					</div>

					<div class="flex items-center gap-2 shrink-0">
						{#if flatSteps.length > 0}
							{@const nextIdx = flatSteps.findIndex(f => f.step.status === 'in_progress' || f.step.status === 'pending')}
							{@const targetId = nextIdx !== -1 ? flatSteps[nextIdx].step.id : flatSteps[0].step.id}
							<button
								type="button"
								onclick={() => navigateToStep(targetId)}
								class="inline-flex items-center gap-1.5 rounded-xl bg-indigo-600 px-4 py-2 text-xs font-medium text-white shadow-xs hover:bg-indigo-500"
							>
								<Play class="h-3.5 w-3.5 fill-current" />
								<span>Start Next Item</span>
							</button>
						{/if}

						<button
							type="button"
							onclick={handleDuplicate}
							class="p-2 rounded-xl border border-zinc-800 bg-zinc-900 text-zinc-400 hover:text-white"
							title="Duplicate"
						>
							<Copy class="h-3.5 w-3.5" />
						</button>
						<button
							type="button"
							onclick={handleDeleteSyllabus}
							class="p-2 rounded-xl border border-zinc-800 bg-zinc-900 text-zinc-400 hover:text-red-400"
							title="Delete"
						>
							<Trash2 class="h-3.5 w-3.5" />
						</button>
					</div>
				</div>

				<!-- Progress -->
				<div class="pt-3 border-t border-zinc-800/80 space-y-1.5">
					<div class="flex items-center justify-between text-xs font-mono">
						<span class="text-zinc-400">{doneCount} of {totalCount} items done</span>
						<span class="font-bold {progressPct === 100 ? 'text-emerald-400' : 'text-indigo-400'}">{progressPct}%</span>
					</div>
					<div class="h-1.5 w-full rounded-full bg-zinc-800 overflow-hidden">
						<div
							class="h-full rounded-full transition-all duration-300 {progressPct === 100 ? 'bg-emerald-500' : 'bg-indigo-500'}"
							style="width: {progressPct}%"
						></div>
					</div>
				</div>
			</div>

			<!-- Sections / Queue Categories -->
			<div class="space-y-4">
				<div class="flex items-center justify-between">
					<h2 class="text-sm font-bold uppercase tracking-wider text-zinc-300">Sections</h2>
					<button
						type="button"
						onclick={openCreateModule}
						class="inline-flex items-center gap-1 rounded-lg bg-zinc-800 px-3 py-1.5 text-xs text-zinc-200 hover:bg-zinc-700"
					>
						<Plus class="h-3.5 w-3.5" />
						<span>Add Section</span>
					</button>
				</div>

				{#if !syllabus.modules || syllabus.modules.length === 0}
					<div class="rounded-xl border border-dashed border-zinc-800 p-8 text-center text-xs text-zinc-500">
						No sections yet. Click "Add Section" to group what you want to learn.
					</div>
				{:else}
					<div class="space-y-5">
						{#each syllabus.modules as mod (mod.id)}
							<div class="rounded-xl border border-zinc-800 bg-zinc-900/30 overflow-hidden">
								<!-- Section Bar -->
								<div class="flex items-center justify-between px-4 py-2.5 bg-zinc-900/60 border-b border-zinc-800/60">
									<h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">{mod.title}</h3>

									<div class="flex items-center gap-1">
										<button
											type="button"
											onclick={() => openCreateStep(mod.id)}
											class="p-1 text-zinc-400 hover:text-white"
											title="Add item with notes"
										>
											<Plus class="h-3.5 w-3.5" />
										</button>
										<button
											type="button"
											onclick={() => openEditModule(mod)}
											class="p-1 text-zinc-400 hover:text-white"
										>
											<Edit3 class="h-3.5 w-3.5" />
										</button>
										<button
											type="button"
											onclick={() => handleDeleteModule(mod.id, mod.title)}
											class="p-1 text-zinc-400 hover:text-red-400"
										>
											<Trash2 class="h-3.5 w-3.5" />
										</button>
									</div>
								</div>

								<!-- Items List -->
								<div class="divide-y divide-zinc-800/40">
									{#if !mod.steps || mod.steps.length === 0}
										<div class="p-4 text-center text-xs text-zinc-500">No items yet. Type below to add to queue.</div>
									{:else}
										{#each mod.steps as step (step.id)}
											<div class="flex items-center justify-between px-4 py-2.5 hover:bg-zinc-800/20 text-xs">
												<button
													type="button"
													onclick={() => navigateToStep(step.id)}
													class="flex items-center gap-2.5 text-left flex-1 truncate hover:text-indigo-400"
												>
													{#if step.status === 'completed'}
														<CheckCircle2 class="h-3.5 w-3.5 text-emerald-400 shrink-0" />
													{:else if step.status === 'in_progress'}
														<div class="h-2 w-2 rounded-full bg-indigo-500 shrink-0"></div>
													{:else if step.status === 'skipped'}
														<FastForward class="h-3 w-3 text-amber-400 shrink-0" />
													{:else}
														<Circle class="h-3 w-3 text-zinc-500 shrink-0" />
													{/if}
													<span class="truncate font-medium text-zinc-200">{step.title}</span>

													{#if $activeFocusStepIds.has(step.id)}
														<span class="inline-flex items-center gap-1 rounded bg-indigo-500/15 border border-indigo-500/30 px-1.5 py-0.2 text-[10px] text-indigo-300 font-mono shrink-0">
															<Target class="h-2.5 w-2.5" />
															<span>In Next</span>
														</span>
													{/if}
												</button>

												<div class="flex items-center gap-2 shrink-0">
													{#if (step.resources && step.resources.length > 0) || step.resource_url}
														{@const primaryLink = (step.resources && step.resources.length > 0) ? step.resources[0] : step.resource_url}
														<a
															href={primaryLink}
															target="_blank"
															rel="noopener noreferrer"
															class="p-1 text-zinc-400 hover:text-indigo-400"
															title={step.resources && step.resources.length > 1 ? `${step.resources.length} resource links` : 'Open link'}
														>
															<ExternalLink class="h-3 w-3" />
														</a>
													{/if}

													<button
														type="button"
														onclick={() => handleToggleFocus(step.id)}
														class="p-1 rounded transition-colors {$activeFocusStepIds.has(step.id) ? 'text-indigo-400 bg-indigo-500/15 hover:bg-indigo-500/25' : 'text-zinc-500 hover:text-indigo-400'}"
														title={$activeFocusStepIds.has(step.id) ? 'In What Next (click to remove)' : 'Add to What Next'}
													>
														<Target class="h-3.5 w-3.5" />
													</button>

													<button
														type="button"
														onclick={() => openEditStep(step)}
														class="p-1 text-zinc-500 hover:text-zinc-300"
													>
														<Edit3 class="h-3 w-3" />
													</button>
													<button
														type="button"
														onclick={() => handleDeleteStep(step.id, step.title)}
														class="p-1 text-zinc-500 hover:text-red-400"
													>
														<Trash2 class="h-3 w-3" />
													</button>
												</div>
											</div>
										{/each}
									{/if}

									<!-- Quick Add Input -->
									<div class="p-2.5 bg-zinc-950/40">
										<form 
											onsubmit={(e) => { e.preventDefault(); handleQuickAdd(mod.id); }}
											class="flex items-center gap-2"
										>
											<input
												bind:this={quickAddInputRefs[mod.id]}
												type="text"
												bind:value={quickAddInputs[mod.id]}
												placeholder="+ Add book, article, paper, video, or topic to queue (press Enter)..."
												class="flex-1 bg-transparent border-0 px-2 py-1 text-xs text-zinc-200 placeholder-zinc-500 focus:outline-hidden"
											/>
											<button
												type="submit"
												disabled={!quickAddInputs[mod.id]?.trim() || quickAddLoading[mod.id]}
												class="inline-flex items-center gap-1 rounded-md bg-zinc-800 px-2.5 py-1 text-[11px] font-medium text-zinc-300 hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed"
											>
												<CornerDownLeft class="h-3 w-3" />
												<span>Add</span>
											</button>
										</form>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Edit List Modal -->
{#if showSyllabusModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl border border-zinc-700 bg-zinc-900 p-6 shadow-2xl">
			<h3 class="text-base font-semibold text-zinc-100">Edit Topic / List Details</h3>
			<form onsubmit={handleSaveSyllabus} class="mt-4 space-y-4">
				<div>
					<label for="edit-s-title" class="block text-xs font-medium text-zinc-300 mb-1">Topic / List Title</label>
					<input
						id="edit-s-title"
						type="text"
						bind:value={sylTitle}
						required
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 focus:border-indigo-500 focus:outline-hidden"
					/>
				</div>
				<div>
					<label for="edit-s-desc" class="block text-xs font-medium text-zinc-300 mb-1">Description</label>
					<textarea
						id="edit-s-desc"
						bind:value={sylDesc}
						rows="3"
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 focus:border-indigo-500 focus:outline-hidden"
					></textarea>
				</div>
				<div class="flex items-center justify-end gap-2 pt-2">
					<button type="button" onclick={() => showSyllabusModal = false} class="px-3 py-1.5 text-xs text-zinc-400">Cancel</button>
					<button type="submit" class="rounded-xl bg-indigo-600 px-4 py-2 text-xs font-medium text-white hover:bg-indigo-500">Save</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Add / Edit Section Modal -->
{#if showModuleModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl border border-zinc-700 bg-zinc-900 p-6 shadow-2xl">
			<h3 class="text-base font-semibold text-zinc-100">{editingModule ? 'Edit Section' : 'Add Section'}</h3>
			<form onsubmit={handleSaveModule} class="mt-4 space-y-4">
				<div>
					<label for="m-title" class="block text-xs font-medium text-zinc-300 mb-1">Section Title</label>
					<input
						id="m-title"
						type="text"
						bind:value={moduleTitle}
						placeholder="e.g. Currently Reading or Systems Backlog"
						required
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 focus:border-indigo-500 focus:outline-hidden"
					/>
				</div>
				<div class="flex items-center justify-end gap-2 pt-2">
					<button type="button" onclick={() => showModuleModal = false} class="px-3 py-1.5 text-xs text-zinc-400">Cancel</button>
					<button type="submit" disabled={isSavingModule} class="rounded-xl bg-indigo-600 px-4 py-2 text-xs font-medium text-white hover:bg-indigo-500">
						{isSavingModule ? 'Saving...' : 'Save'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Detailed Add / Edit Item Modal (No fake criteria, just title, multiple links, context) -->
{#if showStepModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-xs">
		<div class="w-full max-w-lg rounded-2xl border border-zinc-700 bg-zinc-900 p-6 shadow-2xl">
			<h3 class="text-base font-semibold text-zinc-100">{editingStep ? 'Edit Item' : 'New Item'}</h3>
			<form onsubmit={handleSaveStep} class="mt-4 space-y-3.5">
				<div>
					<label for="st-title" class="block text-xs font-medium text-zinc-300 mb-1">Title / Book / Topic</label>
					<input
						id="st-title"
						type="text"
						bind:value={stepTitle}
						placeholder="e.g. Designing Data Intensive Applications Ch. 5"
						required
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 focus:border-indigo-500 focus:outline-hidden"
					/>
				</div>

				<!-- Unlimited Multiple Links / Resources -->
				<div class="space-y-1.5">
					<div class="flex items-center justify-between">
						<span class="block text-xs font-medium text-zinc-300">
							Resource Links (Unlimited)
						</span>
						<button
							type="button"
							onclick={addResourceLink}
							class="inline-flex items-center gap-1 text-[11px] font-medium text-indigo-400 hover:text-indigo-300"
						>
							<Plus class="h-3 w-3" />
							<span>Add another link</span>
						</button>
					</div>

					<div class="space-y-2 max-h-48 overflow-y-auto pr-1">
						{#each stepResources as link, idx}
							<div class="flex items-center gap-2">
								<span class="text-zinc-500 font-mono text-[11px] shrink-0">Link:</span>
								<input
									type="url"
									bind:value={stepResources[idx]}
									placeholder="https://..."
									class="flex-1 rounded-xl border border-zinc-700 bg-zinc-800 px-3 py-1.5 font-mono text-xs text-zinc-100 focus:border-indigo-500 focus:outline-hidden"
								/>
								{#if stepResources.length > 1}
									<button
										type="button"
										onclick={() => removeResourceLink(idx)}
										class="p-1 text-zinc-500 hover:text-red-400"
										title="Remove link"
									>
										<X class="h-3.5 w-3.5" />
									</button>
								{/if}
							</div>
						{/each}
					</div>
				</div>

				<div>
					<label for="st-content" class="block text-xs font-medium text-zinc-300 mb-1">Context / Thoughts (optional)</label>
					<textarea
						id="st-content"
						bind:value={stepContent}
						rows="3"
						placeholder="Why I'm learning this, chapters to focus on, or questions..."
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800 p-2.5 font-mono text-xs text-zinc-100 focus:border-indigo-500 focus:outline-hidden"
					></textarea>
				</div>

				<div class="flex items-center justify-end gap-2 pt-2">
					<button type="button" onclick={() => showStepModal = false} class="px-3 py-1.5 text-xs text-zinc-400">Cancel</button>
					<button type="submit" disabled={isSavingStep} class="rounded-xl bg-indigo-600 px-4 py-2 text-xs font-medium text-white hover:bg-indigo-500">
						{isSavingStep ? 'Saving...' : 'Save Item'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
