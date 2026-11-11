<script lang="ts">
	import { onMount } from 'svelte';
	import { 
		api, 
		type FocusItem, 
		type BacklogStepItem,
		type FocusStatus 
	} from '#lib/api';
	import { focusList, refreshFocus } from '#lib/focusStore';
	import { refreshSyllabi } from '#lib/syllabusStore';
	import { 
		Target, 
		Plus, 
		CheckCircle2, 
		Circle, 
		ExternalLink, 
		BookOpen, 
		Search, 
		X, 
		FileText, 
		ListPlus, 
		Check, 
		GripVertical 
	} from '@lucide/svelte';

	let activeTab = $state<'next' | 'log'>('next');
	let loading = $state(true);
	let error = $state('');

	// Picker modal state
	let showPickerModal = $state(false);
	let backlogItems = $state<BacklogStepItem[]>([]);
	let pickerSearch = $state('');
	let loadingBacklog = $state(false);

	// Quick add state
	let showQuickAdd = $state(false);
	let quickTitle = $state('');
	let quickUrl = $state('');
	let isQuickAdding = $state(false);

	// Notes editing tracking
	let editingNotesId = $state<string | null>(null);
	let notesDraft = $state<Record<string, string>>({});

	onMount(async () => {
		try {
			loading = true;
			await refreshFocus();
		} catch (err: any) {
			error = err.message || 'Failed to load list';
		} finally {
			loading = false;
		}
	});

	// Filtered items based on tab
	let nextItems = $derived($focusList.filter(item => item.status !== 'completed'));
	let completedLog = $derived($focusList.filter(item => item.status === 'completed'));

	let displayedItems = $derived(
		activeTab === 'next' ? nextItems : completedLog
	);

	async function openPicker() {
		showPickerModal = true;
		pickerSearch = '';
		loadingBacklog = true;
		try {
			backlogItems = await api.listBacklogItems();
		} catch (err: any) {
			alert(err.message || 'Failed to load backlog items');
		} finally {
			loadingBacklog = false;
		}
	}

	let filteredBacklog = $derived(
		backlogItems.filter(item => {
			if (!pickerSearch.trim()) return true;
			const q = pickerSearch.toLowerCase();
			return item.title.toLowerCase().includes(q) ||
				item.module_title.toLowerCase().includes(q) ||
				item.syllabus_title.toLowerCase().includes(q);
		})
	);

	async function addFromBacklog(item: BacklogStepItem) {
		try {
			await api.createFocusItem({
				step_id: item.id,
				period: 'next'
			});
			backlogItems = backlogItems.map(b => b.id === item.id ? { ...b, in_focus: true } : b);
			await refreshFocus();
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to add item');
		}
	}

	async function handleQuickAdd(e: Event) {
		e.preventDefault();
		const cleanedTitle = quickTitle
			.replace(/^[\s\-–—*•]+/, '')
			.replace(/[\s\-–—*•]+$/, '')
			.replace(/[\-–—]{2,}/g, ' ')
			.replace(/\s+/g, ' ')
			.trim();
		if (!cleanedTitle) return;
		isQuickAdding = true;
		try {
			const urls = quickUrl.match(/https?:\/\/[^\s,]+/g) || (quickUrl.trim() ? [quickUrl.trim()] : []);
			await api.createFocusItem({
				custom_title: cleanedTitle,
				resource_url: urls[0] || '',
				resources: urls,
				period: 'next'
			});
			quickTitle = '';
			quickUrl = '';
			showQuickAdd = false;
			await refreshFocus();
		} catch (err: any) {
			alert(err.message || 'Failed to add item');
		} finally {
			isQuickAdding = false;
		}
	}

	async function updateStatus(item: FocusItem, status: FocusStatus) {
		try {
			await api.updateFocusItem(item.id, { status });
			await refreshFocus();
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to update status');
		}
	}

	async function saveNotes(item: FocusItem) {
		const text = notesDraft[item.id] ?? item.notes;
		try {
			await api.updateFocusItem(item.id, { notes: text });
			editingNotesId = null;
			await refreshFocus();
		} catch (err: any) {
			alert(err.message || 'Failed to save notes');
		}
	}

	async function removeFromWhatNext(id: string) {
		try {
			await api.deleteFocusItem(id);
			await refreshFocus();
		} catch (err: any) {
			alert(err.message || 'Failed to remove');
		}
	}

	// Drag-and-drop state & handlers
	let draggedIndex = $state<number | null>(null);
	let dragOverIndex = $state<number | null>(null);

	function handleDragStart(e: DragEvent, index: number) {
		if (activeTab !== 'next') return;
		draggedIndex = index;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
			e.dataTransfer.setData('text/plain', index.toString());
		}
	}

	function handleDragOver(e: DragEvent, index: number) {
		if (activeTab !== 'next' || draggedIndex === null) return;
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'move';
		}
		dragOverIndex = index;
	}

	function handleDragEnter(e: DragEvent, index: number) {
		if (activeTab !== 'next' || draggedIndex === null) return;
		dragOverIndex = index;
	}

	function handleDragEnd() {
		draggedIndex = null;
		dragOverIndex = null;
	}

	async function handleDrop(e: DragEvent, targetIndex: number) {
		if (activeTab !== 'next' || draggedIndex === null) return;
		e.preventDefault();
		const fromIndex = draggedIndex;
		draggedIndex = null;
		dragOverIndex = null;

		if (fromIndex === targetIndex) return;

		const reordered = [...nextItems];
		const [moved] = reordered.splice(fromIndex, 1);
		reordered.splice(targetIndex, 0, moved);

		// Optimistically update store
		focusList.update(current => {
			const completed = current.filter(i => i.status === 'completed');
			return [...reordered, ...completed];
		});

		try {
			await api.reorderFocusItems(reordered.map(i => i.id));
		} catch (err: any) {
			console.error('Failed to save reordered items', err);
			await refreshFocus();
		}
	}
</script>

<div class="max-w-5xl mx-auto px-6 py-6 space-y-6">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-800">
		<div>
			<div class="flex items-center gap-2">
				<div class="flex h-7 w-7 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
					<Target class="h-4 w-4" />
				</div>
				<h1 class="text-xl font-bold tracking-tight text-white">What Next</h1>
			</div>
			<p class="mt-1 text-xs text-zinc-400">
				Whenever you sit down to learn, this is what's queued up next.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<button
				type="button"
				onclick={() => showQuickAdd = !showQuickAdd}
				class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-xs font-medium text-zinc-300 hover:text-white hover:bg-zinc-800 transition-colors"
			>
				<Plus class="h-3.5 w-3.5" />
				<span>Quick Target</span>
			</button>

			<button
				type="button"
				onclick={openPicker}
				class="inline-flex items-center gap-1.5 rounded-lg bg-indigo-600 px-3.5 py-1.5 text-xs font-medium text-white hover:bg-indigo-500 transition-colors shadow-xs"
			>
				<ListPlus class="h-3.5 w-3.5" />
				<span>Pick from Backlog</span>
			</button>
		</div>
	</div>

	<!-- Quick Target Drawer -->
	{#if showQuickAdd}
		<form onsubmit={handleQuickAdd} class="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4 space-y-3">
			<div class="flex items-center justify-between">
				<span class="text-xs font-medium text-zinc-200">
					Add Quick Target to What Next
				</span>
				<button type="button" onclick={() => showQuickAdd = false} class="text-zinc-500 hover:text-zinc-300">
					<X class="h-4 w-4" />
				</button>
			</div>
			<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
				<input
					type="text"
					bind:value={quickTitle}
					placeholder="What do you want to learn / read / build next?"
					required
					class="sm:col-span-2 rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-xs text-zinc-100 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden"
				/>
				<input
					type="url"
					bind:value={quickUrl}
					placeholder="URL / article link (optional)"
					class="rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-xs text-zinc-100 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden"
				/>
			</div>
			<div class="flex justify-end gap-2">
				<button
					type="button"
					onclick={() => showQuickAdd = false}
					class="px-2.5 py-1 text-xs text-zinc-400 hover:text-zinc-200"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={isQuickAdding || !quickTitle.trim()}
					class="rounded-lg bg-indigo-600 px-3 py-1 text-xs font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
				>
					{isQuickAdding ? 'Adding...' : 'Add to What Next'}
				</button>
			</div>
		</form>
	{/if}

	<!-- View Tabs -->
	<div class="flex items-center justify-between border-b border-zinc-800 pb-2">
		<div class="flex items-center gap-1.5">
			<button
				type="button"
				onclick={() => activeTab = 'next'}
				class="inline-flex items-center gap-2 rounded-lg px-3 py-1.5 text-xs font-medium transition-colors {activeTab === 'next' ? 'bg-zinc-800 text-white font-semibold' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-850'}"
			>
				<span>What Next</span>
				{#if nextItems.length > 0}
					<span class="rounded-full bg-indigo-500/20 text-indigo-300 px-1.5 py-0.2 text-[10px] font-mono font-semibold">
						{nextItems.length}
					</span>
				{/if}
			</button>

			<button
				type="button"
				onclick={() => activeTab = 'log'}
				class="inline-flex items-center gap-2 rounded-lg px-3 py-1.5 text-xs font-medium transition-colors {activeTab === 'log' ? 'bg-zinc-800 text-white font-semibold' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-850'}"
			>
				<span>Finished Log</span>
				{#if completedLog.length > 0}
					<span class="rounded-full bg-emerald-500/20 text-emerald-400 px-1.5 py-0.2 text-[10px] font-mono">
						{completedLog.length}
					</span>
				{/if}
			</button>
		</div>

		<span class="text-[11px] text-zinc-500 hidden sm:inline">
			{activeTab === 'next' ? (nextItems.length > 1 ? 'Drag items to reorder priority' : 'Items queued up for your next study session') : 'Completed items & learning takeaways'}
		</span>
	</div>

	<!-- Content Area -->
	{#if loading}
		<div class="rounded-xl border border-zinc-800 bg-zinc-900/20 divide-y divide-zinc-800/50 animate-pulse">
			{#each Array(3) as _}
				<div class="h-16 px-4 py-3 flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="h-6 w-6 rounded-full bg-zinc-800"></div>
						<div class="space-y-1.5">
							<div class="h-3.5 w-44 rounded bg-zinc-800"></div>
							<div class="h-2.5 w-64 rounded bg-zinc-800/60"></div>
						</div>
					</div>
					<div class="h-6 w-20 rounded bg-zinc-800"></div>
				</div>
			{/each}
		</div>
	{:else if error}
		<div class="rounded-xl border border-red-500/20 bg-red-500/10 p-6 text-center text-xs text-red-300">
			{error}
		</div>
	{:else if displayedItems.length === 0}
		<!-- Empty state for current tab -->
		<div class="rounded-2xl border border-dashed border-zinc-800 p-10 text-center max-w-md mx-auto space-y-3">
			<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-zinc-900 border border-zinc-800 text-zinc-500 mx-auto">
				{#if activeTab === 'log'}
					<CheckCircle2 class="h-5 w-5" />
				{:else}
					<Target class="h-5 w-5" />
				{/if}
			</div>
			<h2 class="text-sm font-semibold text-zinc-200">
				{activeTab === 'next' ? 'Nothing Queued Up Next' : 'No Finished Items Yet'}
			</h2>
			<p class="text-xs text-zinc-400 leading-relaxed">
				{activeTab === 'log' 
					? 'Finished items and your study notes will appear here.' 
					: 'Pick a few things from your queues that you want to tackle next.'}
			</p>
			{#if activeTab === 'next'}
				<div class="pt-2">
					<button
						type="button"
						onclick={openPicker}
						class="inline-flex items-center gap-1.5 rounded-lg bg-indigo-600 px-3.5 py-1.5 text-xs font-medium text-white hover:bg-indigo-500 transition-colors"
					>
						<ListPlus class="h-3.5 w-3.5" />
						<span>Pick from Backlog</span>
					</button>
				</div>
			{/if}
		</div>
	{:else}
		<!-- List of Items -->
		<div role="list" class="rounded-xl border border-zinc-800 bg-zinc-900/30 divide-y divide-zinc-800/70 overflow-hidden">
			{#each displayedItems as item, index (item.id)}
				<div 
					role="listitem"
					class="p-4 hover:bg-zinc-850 hover:bg-zinc-800/30 transition-all space-y-3 {draggedIndex === index ? 'opacity-35 bg-indigo-500/10' : ''} {dragOverIndex === index && draggedIndex !== index ? 'border-t-2 border-t-indigo-500 bg-indigo-500/10' : ''}"
					draggable={activeTab === 'next'}
					ondragstart={(e) => handleDragStart(e, index)}
					ondragover={(e) => handleDragOver(e, index)}
					ondragenter={(e) => handleDragEnter(e, index)}
					ondragend={handleDragEnd}
					ondrop={(e) => handleDrop(e, index)}
				>
					<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
						<!-- Left: drag handle + status + title + breadcrumbs -->
						<div class="flex items-start gap-2.5 min-w-0 flex-1">
							{#if activeTab === 'next'}
								<div 
									class="pt-1 text-zinc-600 hover:text-zinc-300 cursor-grab active:cursor-grabbing shrink-0 transition-colors"
									title="Drag to reorder"
								>
									<GripVertical class="h-3.5 w-3.5" />
								</div>
							{/if}

							<!-- Status cycle button -->
							<div class="pt-0.5 shrink-0">
								{#if item.status === 'completed'}
									<button
										type="button"
										onclick={() => updateStatus(item, 'todo')}
										class="text-emerald-400 hover:text-emerald-300"
										title="Mark incomplete"
									>
										<CheckCircle2 class="h-4 w-4" />
									</button>
								{:else if item.status === 'in_progress'}
									<button
										type="button"
										onclick={() => updateStatus(item, 'completed')}
										class="h-4 w-4 rounded-full border-2 border-indigo-400 flex items-center justify-center group"
										title="Click to mark Done"
									>
										<div class="h-2 w-2 rounded-full bg-indigo-400"></div>
									</button>
								{:else}
									<button
										type="button"
										onclick={() => updateStatus(item, 'in_progress')}
										class="text-zinc-500 hover:text-indigo-400 transition-colors"
										title="Click to start Doing"
									>
										<Circle class="h-4 w-4" />
									</button>
								{/if}
							</div>

							<div class="min-w-0 flex-1">
								<div class="flex flex-wrap items-center gap-2">
									{#if item.step_id && item.syllabus_id}
										<a
											href="/syllabus/{item.syllabus_id}?step={item.step_id}"
											class="text-sm font-medium text-zinc-100 hover:text-indigo-300 hover:underline transition-colors {item.status === 'completed' ? 'line-through text-zinc-500' : ''}"
										>
											{item.title}
										</a>
									{:else}
										<span class="text-sm font-medium text-zinc-100 {item.status === 'completed' ? 'line-through text-zinc-500' : ''}">
											{item.title}
										</span>
									{/if}

									<!-- Source list and section badge -->
									{#if item.syllabus_title}
										<span class="inline-flex items-center gap-1 rounded-md bg-zinc-800 px-2 py-0.5 text-[10px] text-zinc-400 font-mono">
											<BookOpen class="h-2.5 w-2.5" />
											<span>{item.syllabus_title}</span>
											{#if item.module_title}
												<span class="text-zinc-600">›</span>
												<span>{item.module_title}</span>
											{/if}
										</span>
									{/if}
								</div>

								<!-- Resource link display (plain url + hyperlink for unlimited links) -->
								{#if (item.resources && item.resources.length > 0) || item.resource_url}
									{@const allLinks = (item.resources && item.resources.length > 0) ? item.resources : [item.resource_url]}
									<div class="mt-1.5 space-y-1">
										{#each allLinks as link}
											<div class="flex items-center gap-1.5 text-xs text-zinc-400">
												<span class="text-zinc-500 font-mono text-[11px] shrink-0">Link:</span>
												<a
													href={link}
													target="_blank"
													rel="noopener noreferrer"
													class="text-indigo-400 hover:text-indigo-300 hover:underline inline-flex items-center gap-1 text-xs truncate max-w-full"
												>
													<span class="truncate">{link}</span>
													<ExternalLink class="h-3 w-3 shrink-0" />
												</a>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						</div>

						<!-- Right: Status pills & actions -->
						<div class="flex items-center gap-2 shrink-0 self-end sm:self-center">
							<!-- Status pills -->
							<div class="inline-flex rounded-lg bg-zinc-800/80 p-0.5 text-[11px]">
								<button
									type="button"
									onclick={() => updateStatus(item, 'todo')}
									class="px-2 py-0.5 rounded-md transition-colors {item.status === 'todo' ? 'bg-zinc-700 text-white font-medium' : 'text-zinc-400 hover:text-zinc-200'}"
								>
									To Do
								</button>
								<button
									type="button"
									onclick={() => updateStatus(item, 'in_progress')}
									class="px-2 py-0.5 rounded-md transition-colors {item.status === 'in_progress' ? 'bg-indigo-600 text-white font-medium' : 'text-zinc-400 hover:text-zinc-200'}"
								>
									Doing
								</button>
								<button
									type="button"
									onclick={() => updateStatus(item, 'completed')}
									class="px-2 py-0.5 rounded-md transition-colors {item.status === 'completed' ? 'bg-emerald-600 text-white font-medium' : 'text-zinc-400 hover:text-zinc-200'}"
								>
									Done
								</button>
							</div>

							<!-- Toggle notes scratchpad -->
							<button
								type="button"
								onclick={() => {
									if (editingNotesId === item.id) {
										editingNotesId = null;
									} else {
										editingNotesId = item.id;
										notesDraft[item.id] = item.notes;
									}
								}}
								class="p-1.5 rounded-lg text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 transition-colors {item.notes ? 'text-indigo-400' : ''}"
								title="Notes / Takeaways"
							>
								<FileText class="h-3.5 w-3.5" />
							</button>

							<!-- Remove from What Next -->
							<button
								type="button"
								onclick={() => removeFromWhatNext(item.id)}
								class="p-1.5 rounded-lg text-zinc-500 hover:text-zinc-200 hover:bg-zinc-800 transition-colors"
								title="Remove from What Next"
							>
								<X class="h-3.5 w-3.5" />
							</button>
						</div>
					</div>

					<!-- Notes Drawer / Log entry -->
					{#if editingNotesId === item.id || item.notes}
						<div class="ml-7 pt-2 border-t border-zinc-800/60">
							{#if editingNotesId === item.id}
								<div class="space-y-2">
									<textarea
										bind:value={notesDraft[item.id]}
										rows="3"
										placeholder="Jot down notes, reflections, key insights, or questions..."
										class="w-full rounded-lg border border-zinc-700 bg-zinc-800/80 p-2 text-xs text-zinc-200 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden"
									></textarea>
									<div class="flex items-center justify-end gap-2">
										<button
											type="button"
											onclick={() => editingNotesId = null}
											class="px-2.5 py-1 text-[11px] text-zinc-400 hover:text-zinc-200"
										>
											Cancel
										</button>
										<button
											type="button"
											onclick={() => saveNotes(item)}
											class="inline-flex items-center gap-1 rounded-md bg-indigo-600 px-2.5 py-1 text-[11px] font-medium text-white hover:bg-indigo-500"
										>
											<Check class="h-3 w-3" />
											<span>Save Notes</span>
										</button>
									</div>
								</div>
							{:else}
								<div 
									class="rounded-lg bg-zinc-800/40 p-2.5 text-xs text-zinc-300 font-mono cursor-pointer hover:bg-zinc-800/70 transition-colors"
									onclick={() => {
										editingNotesId = item.id;
										notesDraft[item.id] = item.notes;
									}}
									role="button"
									tabindex="0"
									onkeydown={(e) => {
										if (e.key === 'Enter') {
											editingNotesId = item.id;
											notesDraft[item.id] = item.notes;
										}
									}}
									title="Click to edit notes"
								>
									<div class="flex items-center justify-between text-[10px] text-zinc-500 mb-1">
										<span>Notes</span>
										<span class="hover:text-indigo-400">Edit</span>
									</div>
									<div class="whitespace-pre-wrap">{item.notes}</div>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Backlog Picker Modal -->
{#if showPickerModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-xs">
		<div class="w-full max-w-2xl rounded-2xl border border-zinc-700 bg-zinc-900 p-6 shadow-2xl flex flex-col max-h-[85vh]">
			<div class="flex items-center justify-between pb-3 border-b border-zinc-800">
				<div>
					<h3 class="text-base font-semibold text-zinc-100">Pick From Learning Queues</h3>
					<p class="text-xs text-zinc-400 mt-0.5">Select items to put into your What Next list</p>
				</div>
				<button
					type="button"
					onclick={() => showPickerModal = false}
					class="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800"
				>
					<X class="h-4 w-4" />
				</button>
			</div>

			<!-- Search filter -->
			<div class="pt-4 pb-3">
				<div class="relative">
					<Search class="h-3.5 w-3.5 text-zinc-500 absolute left-3 top-2.5" />
					<input
						type="text"
						bind:value={pickerSearch}
						placeholder="Search by title, topic, or section..."
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800 pl-9 pr-3 py-1.5 text-xs text-zinc-100 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden"
					/>
				</div>
			</div>

			<!-- Items list -->
			<div class="flex-1 overflow-y-auto space-y-2 py-2 pr-1">
				{#if loadingBacklog}
					<div class="p-8 text-center text-xs text-zinc-400">Loading backlog items...</div>
				{:else if filteredBacklog.length === 0}
					<div class="p-8 text-center text-xs text-zinc-500">
						No matching items found.
					</div>
				{:else}
					{#each filteredBacklog as step (step.id)}
						<div class="flex items-center justify-between gap-3 p-2.5 rounded-xl border border-zinc-800/80 bg-zinc-950/40 hover:bg-zinc-800/40 transition-colors">
							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-2">
									<span class="text-xs font-medium text-zinc-200 truncate">{step.title}</span>
									{#if step.status === 'completed'}
										<span class="rounded bg-emerald-500/10 px-1.5 py-0.2 text-[9px] text-emerald-400 font-mono">Done</span>
									{:else if step.status === 'in_progress'}
										<span class="rounded bg-indigo-500/10 px-1.5 py-0.2 text-[9px] text-indigo-400 font-mono">Doing</span>
									{/if}
								</div>
								<div class="flex items-center gap-1.5 text-[10px] text-zinc-500 mt-0.5">
									<span>{step.syllabus_title}</span>
									<span>›</span>
									<span>{step.module_title}</span>
								</div>
							</div>

							<div class="flex items-center gap-1.5 shrink-0">
								{#if step.in_focus}
									<span class="text-[10px] text-indigo-400 font-mono bg-indigo-500/10 px-2 py-1 rounded-md">
										In Next
									</span>
								{:else}
									<button
										type="button"
										onclick={() => addFromBacklog(step)}
										class="px-2.5 py-1 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-[11px] font-medium text-white transition-colors"
									>
										+ What Next
									</button>
								{/if}
							</div>
						</div>
					{/each}
				{/if}
			</div>

			<div class="pt-3 border-t border-zinc-800 flex justify-end">
				<button
					type="button"
					onclick={() => showPickerModal = false}
					class="px-4 py-1.5 rounded-xl bg-zinc-800 hover:bg-zinc-700 text-xs text-zinc-200 font-medium"
				>
					Done
				</button>
			</div>
		</div>
	</div>
{/if}
