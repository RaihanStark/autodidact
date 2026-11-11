<script lang="ts">
	import './layout.css';
	import type { LayoutProps } from './$types';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { checkAuth, isAuthenticated, isAuthModalOpen, lockWorkspace } from '#lib/auth';
	import { 
		syllabiList, 
		activeSyllabus, 
		activeStepId, 
		refreshSyllabi, 
		loadActiveSyllabus,
		sidebarOpen
	} from '#lib/syllabusStore';
	import { refreshFocus, todayFocusCount, activeFocusStepIds } from '#lib/focusStore';
	import { api } from '#lib/api';
	import AuthModal from '#lib/components/AuthModal.svelte';
	import { 
		Compass, 
		Plus, 
		CheckCircle2, 
		Circle, 
		Clock, 
		ShieldCheck, 
		Lock, 
		Menu, 
		X,
		BookOpen,
		FastForward,
		LayoutGrid,
		Target
	} from '@lucide/svelte';

	let { children }: LayoutProps = $props();

	// Modal states
	let showCreateModal = $state(false);
	let newTitle = $state('');
	let newDescription = $state('');
	let isCreating = $state(false);

	// Active route detection
	let currentSyllabusId = $derived(page.params.id || '');
	let currentStepQuery = $derived(page.url.searchParams.get('step') || '');

	onMount(async () => {
		const authed = await checkAuth();
		if (authed) {
			await Promise.all([
				refreshSyllabi(),
				refreshFocus()
			]);
			if (currentSyllabusId) {
				await loadActiveSyllabus(currentSyllabusId);
			}
		}
	});

	$effect(() => {
		if (currentSyllabusId) {
			loadActiveSyllabus(currentSyllabusId);
		}
	});

	$effect(() => {
		activeStepId.set(currentStepQuery);
	});

	async function handleCreateSyllabus(e: Event) {
		e.preventDefault();
		if (!newTitle.trim()) return;
		isCreating = true;
		try {
			const created = await api.createSyllabus({
				title: newTitle.trim(),
				description: newDescription.trim()
			});
			showCreateModal = false;
			newTitle = '';
			newDescription = '';
			await refreshSyllabi();
			goto(`/syllabus/${created.id}`);
		} catch (err: any) {
			alert(err.message || 'Error creating syllabus');
		} finally {
			isCreating = false;
		}
	}
</script>

<svelte:head>
	<title>Autodidact</title>
</svelte:head>

<div class="h-screen w-screen flex bg-zinc-950 text-zinc-100 font-sans overflow-hidden">
	<!-- Left Sidebar -->
	<aside 
		class="fixed inset-y-0 left-0 z-50 flex w-72 flex-col border-r border-zinc-800/80 bg-zinc-900/95 backdrop-blur-md transition-transform duration-200 lg:static lg:translate-x-0 {$sidebarOpen ? 'translate-x-0' : '-translate-x-full'}"
	>
		<!-- Sidebar Top Brand & Quick Actions -->
		<div class="flex h-14 items-center justify-between border-b border-zinc-800/80 px-4">
			<a href="/" class="flex items-center gap-2.5">
				<div class="flex h-7 w-7 items-center justify-center rounded-lg bg-indigo-600 text-white shadow-xs">
					<Compass class="h-4 w-4" />
				</div>
				<span class="font-bold text-sm tracking-tight text-white">Autodidact</span>
			</a>

			<div class="flex items-center gap-1">
				<button
					type="button"
					onclick={() => showCreateModal = true}
					class="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
					title="New List"
				>
					<Plus class="h-4 w-4" />
				</button>
				<button
					type="button"
					onclick={() => sidebarOpen.set(false)}
					class="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 lg:hidden"
				>
					<X class="h-4 w-4" />
				</button>
			</div>
		</div>

		<!-- Learning Queues List & Nested Steps Tree -->
		<div class="flex-1 overflow-y-auto px-3 py-3 space-y-3">
			<div class="space-y-1">
				<a
					href="/"
					class="flex items-center gap-2 rounded-xl px-2.5 py-2 text-xs font-medium transition-colors {page.url.pathname === '/' ? 'bg-indigo-600/15 text-indigo-300 font-semibold border border-indigo-500/20' : 'text-zinc-300 hover:bg-zinc-800/60 hover:text-white'}"
				>
					<LayoutGrid class="h-3.5 w-3.5 shrink-0" />
					<span>Dashboard</span>
				</a>

				<a
					href="/focus"
					class="flex items-center justify-between gap-2 rounded-xl px-2.5 py-2 text-xs font-medium transition-colors {page.url.pathname === '/focus' ? 'bg-indigo-600/15 text-indigo-300 font-semibold border border-indigo-500/20' : 'text-zinc-300 hover:bg-zinc-800/60 hover:text-white'}"
				>
					<div class="flex items-center gap-2">
						<Target class="h-3.5 w-3.5 shrink-0" />
						<span>What Next</span>
					</div>
					{#if $todayFocusCount > 0}
						<span class="rounded-full bg-indigo-500/20 border border-indigo-500/30 px-1.5 py-0.2 text-[10px] font-mono text-indigo-300 font-semibold">
							{$todayFocusCount}
						</span>
					{/if}
				</a>
			</div>

			<div>
				<div class="px-2 mb-2 flex items-center justify-between text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
					<span>Things to Learn</span>
					<span>{$syllabiList.length}</span>
				</div>

				<div class="space-y-1">
					{#each $syllabiList as syl (syl.id)}
						{@const isSelected = syl.id === currentSyllabusId}
						{@const totalCount = syl.total_steps || 0}
						{@const doneCount = syl.completed_steps || 0}
						{@const progressPct = totalCount > 0 ? Math.round((doneCount / totalCount) * 100) : 0}

						<div>
							<a
								href="/syllabus/{syl.id}"
								class="flex items-center justify-between gap-2 rounded-xl px-2.5 py-2 text-xs font-medium transition-colors {isSelected ? 'bg-indigo-600/15 text-indigo-300 font-semibold border border-indigo-500/20' : 'text-zinc-300 hover:bg-zinc-800/60 hover:text-white'}"
							>
								<div class="flex items-center gap-2 truncate">
									<BookOpen class="h-3.5 w-3.5 shrink-0 {isSelected ? 'text-indigo-400' : 'text-zinc-400'}" />
									<span class="truncate">{syl.title}</span>
								</div>
								<span class="font-mono text-[10px] {progressPct === 100 ? 'text-emerald-400' : 'text-zinc-400'} shrink-0">
									{progressPct}%
								</span>
							</a>

							<!-- Steps list rendered directly underneath the selected queue -->
							{#if isSelected && $activeSyllabus && $activeSyllabus.modules}
								<div class="mt-1.5 ml-2.5 pl-2 border-l border-zinc-800 space-y-3 py-1">
									{#each $activeSyllabus.modules as mod, mIdx (mod.id)}
										<div>
											<div class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-zinc-400 truncate">
												{mod.title}
											</div>

											<div class="mt-0.5 space-y-0.5">
												{#each mod.steps as step (step.id)}
													{@const isStepActive = step.id === currentStepQuery}
													<a
														href="/syllabus/{syl.id}?step={step.id}"
														class="flex items-center gap-2 rounded-lg px-2 py-1.5 text-xs transition-colors {isStepActive ? 'bg-zinc-800 text-white font-medium' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40'}"
													>
														{#if step.status === 'completed'}
															<CheckCircle2 class="h-3.5 w-3.5 text-emerald-400 shrink-0" />
														{:else if step.status === 'in_progress'}
															<div class="h-2 w-2 rounded-full bg-indigo-500 shrink-0 ml-0.5 mr-1"></div>
														{:else if step.status === 'skipped'}
															<FastForward class="h-3 w-3 text-amber-400 shrink-0" />
														{:else}
															<Circle class="h-3 w-3 text-zinc-400 shrink-0" />
														{/if}
														<span class="truncate flex-1">{step.title}</span>
														{#if $activeFocusStepIds.has(step.id)}
															<span class="h-1.5 w-1.5 rounded-full bg-indigo-400 shrink-0" title="In What Next"></span>
														{/if}
													</a>
												{/each}
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					{/each}

					{#if $syllabiList.length === 0}
						<div class="p-4 text-center text-xs text-zinc-400">
							No lists yet. Click "+" to start one.
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- Sidebar Bottom Protection Status -->
		<div class="border-t border-zinc-800/80 p-3 flex items-center justify-between text-xs text-zinc-400">
			{#if $isAuthenticated}
				<button
					type="button"
					onclick={() => isAuthModalOpen.set(true)}
					class="inline-flex items-center gap-1.5 text-emerald-400 hover:text-emerald-300 transition-colors"
					title="Local network protection active"
				>
					<ShieldCheck class="h-3.5 w-3.5" />
					<span>Local Protected</span>
				</button>
				<button
					type="button"
					onclick={lockWorkspace}
					class="p-1 rounded text-zinc-400 hover:text-zinc-200"
					title="Lock"
				>
					<Lock class="h-3.5 w-3.5" />
				</button>
			{:else}
				<button
					type="button"
					onclick={() => isAuthModalOpen.set(true)}
					class="inline-flex items-center gap-1.5 text-amber-400 hover:text-amber-300 transition-colors"
				>
					<Lock class="h-3.5 w-3.5" />
					<span>Locked</span>
				</button>
			{/if}
		</div>
	</aside>

	<!-- Mobile Sidebar Overlay Backdrop -->
	{#if $sidebarOpen}
		<button
			type="button"
			class="fixed inset-0 z-40 bg-black/60 backdrop-blur-xs lg:hidden"
			onclick={() => sidebarOpen.set(false)}
			aria-label="Close sidebar"
		></button>
	{/if}

	<!-- Main Workspace Area -->
	<div class="flex-1 flex flex-col h-screen overflow-hidden">
		<!-- Top Bar -->
		<header class="flex h-14 items-center justify-between border-b border-zinc-800/80 bg-zinc-950 px-4 shrink-0">
			<div class="flex items-center gap-3">
				<button
					type="button"
					onclick={() => sidebarOpen.update(v => !v)}
					class="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 lg:hidden"
					aria-label="Toggle sidebar"
				>
					<Menu class="h-5 w-5" />
				</button>

				{#if page.url.pathname === '/'}
					<span class="text-xs font-semibold text-zinc-200">Dashboard</span>
				{:else if $activeSyllabus}
					<div class="flex items-center gap-2 text-xs text-zinc-400 truncate">
						<a href="/" class="hover:text-white text-zinc-400">Dashboard</a>
						<span>/</span>
						<a href="/syllabus/{$activeSyllabus.id}" class="hover:text-white truncate font-medium text-zinc-300">
							{$activeSyllabus.title}
						</a>
						{#if currentStepQuery}
							<span>/</span>
							<span class="text-zinc-400 truncate">Current Item</span>
						{/if}
					</div>
				{/if}
			</div>

			<div></div>
		</header>

		<!-- Main Page Content -->
		<main class="flex-1 overflow-y-auto">
			{@render children()}
		</main>
	</div>

	<!-- Auth Gate Modal -->
	<AuthModal />
</div>

<!-- Create List Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl border border-zinc-700 bg-zinc-900 p-6 shadow-2xl">
			<h3 class="text-base font-semibold text-zinc-100">Things I Want to Learn</h3>
			<p class="text-xs text-zinc-400 mt-1">What topic, technology, or skills do you want to learn next?</p>

			<form onsubmit={handleCreateSyllabus} class="mt-4 space-y-4">
				<div>
					<label for="create-syl-title" class="block text-xs font-medium text-zinc-300 mb-1">Topic / List Title</label>
					<input
						id="create-syl-title"
						type="text"
						bind:value={newTitle}
						placeholder="e.g. Go & Distributed Systems, or Web Dev, or CS50"
						required
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden"
					/>
				</div>
				<div>
					<label for="create-syl-desc" class="block text-xs font-medium text-zinc-300 mb-1">Notes / Why you want to learn this (optional)</label>
					<textarea
						id="create-syl-desc"
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
