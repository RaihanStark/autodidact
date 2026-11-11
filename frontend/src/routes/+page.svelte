<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type Syllabus } from '#lib/api';
	import { syllabiList, refreshSyllabi } from '#lib/syllabusStore';
	import { focusList, refreshFocus } from '#lib/focusStore';
	import { 
		Plus, 
		BookOpen, 
		ArrowRight, 
		Copy, 
		Trash2, 
		CheckCircle2, 
		Layers,
		Sparkles,
		Target
	} from '@lucide/svelte';

	let loading = $state(true);
	let error = $state('');

	// Local modal for creating list
	let showCreateModal = $state(false);
	let newTitle = $state('');
	let newDescription = $state('');
	let isCreating = $state(false);

	onMount(async () => {
		try {
			loading = true;
			await Promise.all([
				refreshSyllabi(),
				refreshFocus()
			]);
		} catch (err: any) {
			error = err.message || 'Failed to load lists';
		} finally {
			loading = false;
		}
	});

	async function handleCreate(e: Event) {
		e.preventDefault();
		if (!newTitle.trim()) return;
		isCreating = true;
		try {
			await api.createSyllabus({
				title: newTitle.trim(),
				description: newDescription.trim()
			});
			showCreateModal = false;
			newTitle = '';
			newDescription = '';
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to create list');
		} finally {
			isCreating = false;
		}
	}

	async function handleDuplicate(id: string) {
		try {
			await api.duplicateSyllabus(id);
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to duplicate');
		}
	}

	async function handleDelete(id: string, title: string) {
		if (!confirm(`Delete "${title}"?`)) return;
		try {
			await api.deleteSyllabus(id);
			await refreshSyllabi();
		} catch (err: any) {
			alert(err.message || 'Failed to delete');
		}
	}

	// Overall stats calculation
	let totalItems = $derived($syllabiList.reduce((acc, s) => acc + (s.total_steps || 0), 0));
	let doneItems = $derived($syllabiList.reduce((acc, s) => acc + (s.completed_steps || 0), 0));
	let overallPct = $derived(totalItems > 0 ? Math.round((doneItems / totalItems) * 100) : 0);
	let whatNextActive = $derived($focusList.filter(item => item.status !== 'completed'));
</script>

<div class="max-w-5xl mx-auto px-6 py-6 space-y-6">
	<!-- Top Header & Actions -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-800">
		<div>
			<h1 class="text-xl font-bold tracking-tight text-white">Things to Learn</h1>
			<p class="mt-0.5 text-xs text-zinc-400">All your active study queues, books, and topics.</p>
		</div>

		<div class="flex items-center gap-2">
			<button
				type="button"
				onclick={() => showCreateModal = true}
				class="inline-flex items-center gap-1.5 rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-indigo-500 transition-colors"
			>
				<Plus class="h-3.5 w-3.5" />
				<span>New List</span>
			</button>
		</div>
	</div>

	{#if loading}
		<div class="rounded-xl border border-zinc-800 bg-zinc-900/20 divide-y divide-zinc-800/50 animate-pulse">
			{#each Array(4) as _}
				<div class="h-14 px-4 py-3 flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="h-8 w-8 rounded-lg bg-zinc-800"></div>
						<div class="space-y-1.5">
							<div class="h-3.5 w-44 rounded bg-zinc-800"></div>
							<div class="h-2.5 w-64 rounded bg-zinc-800/60"></div>
						</div>
					</div>
					<div class="h-3 w-32 rounded bg-zinc-800"></div>
				</div>
			{/each}
		</div>
	{:else if error}
		<div class="rounded-xl border border-red-500/20 bg-red-500/10 p-6 text-center text-xs text-red-300">
			{error}
		</div>
	{:else if $syllabiList.length === 0}
		<!-- Empty State -->
		<div class="rounded-2xl border border-dashed border-zinc-800 p-12 text-center max-w-lg mx-auto space-y-4">
			<div class="flex h-12 w-12 items-center justify-center rounded-xl bg-zinc-900 border border-zinc-800 text-zinc-400 mx-auto">
				<Layers class="h-5 w-5" />
			</div>
			<h2 class="text-sm font-semibold text-zinc-100">No Learning Lists Yet</h2>
			<p class="text-xs text-zinc-400 leading-relaxed">
				Create a list for topics, books, or papers you want to study one by one in order.
			</p>
			<button
				type="button"
				onclick={() => showCreateModal = true}
				class="inline-flex items-center gap-1.5 rounded-xl bg-indigo-600 px-4 py-2 text-xs font-semibold text-white hover:bg-indigo-500"
			>
				<Plus class="h-4 w-4" />
				<span>Create First List</span>
			</button>
		</div>
	{:else}
		<!-- Overall Summary Banner -->
		<div class="rounded-xl border border-zinc-800/90 bg-zinc-900/40 px-4 py-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="flex items-center gap-5 text-xs">
				<div>
					<span class="text-zinc-500 text-[10px] uppercase font-semibold tracking-wider">Lists</span>
					<span class="ml-1.5 text-sm font-bold text-white font-mono">{$syllabiList.length}</span>
				</div>
				<div class="h-4 w-px bg-zinc-800"></div>
				<div>
					<span class="text-zinc-500 text-[10px] uppercase font-semibold tracking-wider">Total Items</span>
					<span class="ml-1.5 text-sm font-bold text-white font-mono">{totalItems}</span>
				</div>
				<div class="h-4 w-px bg-zinc-800"></div>
				<div>
					<span class="text-zinc-500 text-[10px] uppercase font-semibold tracking-wider">Completed</span>
					<span class="ml-1.5 text-sm font-bold text-emerald-400 font-mono">{doneItems}</span>
				</div>
			</div>

			<div class="flex items-center gap-3 w-full sm:w-56">
				<div class="h-1.5 flex-1 bg-zinc-800 rounded-full overflow-hidden">
					<div
						class="h-full rounded-full transition-all duration-300 {overallPct === 100 ? 'bg-emerald-500' : 'bg-indigo-500'}"
						style="width: {overallPct}%"
					></div>
				</div>
				<span class="text-[11px] font-mono text-zinc-400 font-medium shrink-0">{overallPct}% done</span>
			</div>
		</div>

		<!-- What Next Widget (if any active items) -->
		{#if whatNextActive.length > 0}
			<div class="rounded-xl border border-indigo-500/25 bg-indigo-500/5 p-4 space-y-3">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<div class="flex h-6 w-6 items-center justify-center rounded-md bg-indigo-500/20 text-indigo-400">
							<Target class="h-3.5 w-3.5" />
						</div>
						<h2 class="text-xs font-semibold text-zinc-100">What Next</h2>
						<span class="rounded-full bg-indigo-500/20 px-2 py-0.5 text-[10px] font-mono text-indigo-300 font-medium">
							{whatNextActive.length} queued
						</span>
					</div>
					<a
						href="/focus"
						class="text-xs font-medium text-indigo-400 hover:text-indigo-300 flex items-center gap-1 transition-colors"
					>
						<span>Open What Next</span>
						<ArrowRight class="h-3.5 w-3.5" />
					</a>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
					{#each whatNextActive.slice(0, 4) as item (item.id)}
						<a
							href={item.step_id && item.syllabus_id ? `/syllabus/${item.syllabus_id}?step=${item.step_id}` : '/focus'}
							class="flex items-center justify-between gap-2 rounded-lg border border-zinc-800/80 bg-zinc-900/60 px-3 py-2 text-xs hover:border-zinc-700 hover:bg-zinc-800/80 transition-colors"
						>
							<div class="min-w-0 flex-1 truncate">
								<span class="font-medium text-zinc-200 truncate block">{item.title}</span>
								{#if item.syllabus_title}
									<span class="text-[10px] text-zinc-500 truncate block">{item.syllabus_title}</span>
								{/if}
							</div>
							<span class="text-[10px] font-mono px-1.5 py-0.5 rounded {item.status === 'in_progress' ? 'bg-indigo-500/20 text-indigo-300' : 'bg-zinc-800 text-zinc-400'} shrink-0">
								{item.status === 'in_progress' ? 'Doing' : 'To Do'}
							</span>
						</a>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Compact List View -->
		<div class="rounded-xl border border-zinc-800 bg-zinc-900/30 overflow-hidden divide-y divide-zinc-800/70">
			{#each $syllabiList as syl (syl.id)}
				{@const totalCount = syl.total_steps || 0}
				{@const doneCount = syl.completed_steps || 0}
				{@const progressPct = totalCount > 0 ? Math.round((doneCount / totalCount) * 100) : 0}

				<div class="group flex flex-col sm:flex-row sm:items-center justify-between p-3.5 sm:px-4 sm:py-3 hover:bg-zinc-850 hover:bg-zinc-800/40 transition-colors gap-3">
					<!-- Title & Description -->
					<div class="flex items-center gap-3 min-w-0 flex-1">
						<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-zinc-800/80 text-zinc-400 group-hover:text-indigo-400 group-hover:bg-indigo-500/10 transition-colors">
							<BookOpen class="h-4 w-4" />
						</div>
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2">
								<a href="/syllabus/{syl.id}" class="text-sm font-medium text-zinc-200 hover:text-white truncate group-hover:text-indigo-300 transition-colors">
									{syl.title}
								</a>
								{#if progressPct === 100}
									<span class="inline-flex items-center gap-1 rounded-md bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-medium text-emerald-400">
										<CheckCircle2 class="h-3 w-3" />
										Done
									</span>
								{/if}
							</div>
							{#if syl.description}
								<p class="text-xs text-zinc-500 truncate mt-0.5">{syl.description}</p>
							{/if}
						</div>
					</div>

					<!-- Progress & Stats (compact) -->
					<div class="flex items-center gap-4 shrink-0 sm:w-72">
						<div class="flex-1 space-y-1">
							<div class="flex items-center justify-between text-[11px] font-mono">
								<span class="text-zinc-500">{doneCount}/{totalCount} items</span>
								<span class="font-semibold {progressPct === 100 ? 'text-emerald-400' : 'text-zinc-300'}">{progressPct}%</span>
							</div>
							<div class="h-1.5 w-full rounded-full bg-zinc-800 overflow-hidden">
								<div
									class="h-full rounded-full transition-all duration-300 {progressPct === 100 ? 'bg-emerald-500' : 'bg-indigo-500'}"
									style="width: {progressPct}%"
								></div>
							</div>
						</div>

						<!-- Actions -->
						<div class="flex items-center gap-0.5 text-zinc-400">
							<button
								type="button"
								onclick={() => handleDuplicate(syl.id)}
								class="p-1.5 rounded-lg hover:text-zinc-200 hover:bg-zinc-800 transition-colors"
								title="Duplicate"
							>
								<Copy class="h-3.5 w-3.5" />
							</button>
							<button
								type="button"
								onclick={() => handleDelete(syl.id, syl.title)}
								class="p-1.5 rounded-lg hover:text-red-400 hover:bg-red-500/10 transition-colors"
								title="Delete"
							>
								<Trash2 class="h-3.5 w-3.5" />
							</button>
							<a
								href="/syllabus/{syl.id}"
								class="ml-1 p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
								title="Open queue"
							>
								<ArrowRight class="h-4 w-4" />
							</a>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Local Create List Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl border border-zinc-700 bg-zinc-900 p-6 shadow-2xl">
			<h3 class="text-base font-semibold text-zinc-100">Things I Want to Learn</h3>
			<p class="text-xs text-zinc-400 mt-1">What topic, technology, or skills do you want to learn next?</p>

			<form onsubmit={handleCreate} class="mt-4 space-y-4">
				<div>
					<label for="dash-title" class="block text-xs font-medium text-zinc-300 mb-1">Topic / List Title</label>
					<input
						id="dash-title"
						type="text"
						bind:value={newTitle}
						placeholder="e.g. Go & Distributed Systems, or Web Dev, or CS50"
						required
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden"
					/>
				</div>
				<div>
					<label for="dash-desc" class="block text-xs font-medium text-zinc-300 mb-1">Notes / Why you want to learn this (optional)</label>
					<textarea
						id="dash-desc"
						bind:value={newDescription}
						rows="3"
						placeholder="Personal goals, bookmarked links, reference notes..."
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden"
					></textarea>
				</div>
				<div class="flex items-center justify-end gap-2 pt-2">
					<button
						type="button"
						onclick={() => showCreateModal = false}
						class="px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={isCreating || !newTitle.trim()}
						class="rounded-xl bg-indigo-600 px-4 py-2 text-xs font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
					>
						{isCreating ? 'Creating...' : 'Create List'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
