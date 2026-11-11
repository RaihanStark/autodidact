<script lang="ts">
	import { isAuthModalOpen, isAuthenticated, submitKey, currentKey } from '#lib/auth';
	import { Lock, ShieldCheck, Key, AlertCircle } from '@lucide/svelte';

	let inputKey = $state('');
	let errorMsg = $state('');
	let isSubmitting = $state(false);

	$effect(() => {
		currentKey.subscribe(val => {
			if (!inputKey) inputKey = val;
		});
	});

	async function handleUnlock() {
		errorMsg = '';
		isSubmitting = true;
		try {
			const success = await submitKey(inputKey.trim());
			if (!success) {
				errorMsg = 'Invalid security key. Please check your AUTH_TOKEN.';
			}
		} catch (err: any) {
			errorMsg = err.message || 'Verification error occurred';
		} finally {
			isSubmitting = false;
		}
	}

	function useDefaultKey() {
		inputKey = 'autodidact-local-key';
		handleUnlock();
	}
</script>

{#if $isAuthModalOpen}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl border border-zinc-700/80 bg-zinc-900 p-6 shadow-2xl">
			<div class="flex items-center gap-3">
				<div class="flex h-11 w-11 items-center justify-center rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
					<Lock class="h-5 w-5" />
				</div>
				<div>
					<h3 class="text-lg font-semibold text-zinc-100">Local Access Protection</h3>
					<p class="text-xs text-zinc-400">Restricted to your local machine</p>
				</div>
			</div>

			<p class="mt-4 text-sm text-zinc-300 leading-relaxed">
				To prevent unauthorized network access to your learning progress, enter your configured security key.
			</p>

			{#if errorMsg}
				<div class="mt-3 flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-300">
					<AlertCircle class="h-4 w-4 shrink-0 text-red-400" />
					<span>{errorMsg}</span>
				</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); handleUnlock(); }} class="mt-4 space-y-4">
				<div>
					<label for="sec-key" class="block text-xs font-medium text-zinc-400 mb-1.5 flex items-center gap-1.5">
						<Key class="h-3.5 w-3.5 text-zinc-400" />
						Security Key (AUTH_TOKEN)
					</label>
					<input
						id="sec-key"
						type="password"
						bind:value={inputKey}
						placeholder="autodidact-local-key"
						class="w-full rounded-xl border border-zinc-700 bg-zinc-800/80 px-3.5 py-2.5 text-sm text-zinc-100 placeholder-zinc-500 focus:border-indigo-500 focus:outline-hidden focus:ring-1 focus:ring-indigo-500"
					/>
				</div>

				<div class="flex flex-col gap-2 pt-1">
					<button
						type="submit"
						disabled={isSubmitting || !inputKey.trim()}
						class="w-full inline-flex items-center justify-center gap-2 rounded-xl bg-indigo-600 px-4 py-2.5 text-sm font-medium text-white shadow-sm hover:bg-indigo-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<ShieldCheck class="h-4 w-4" />
						{isSubmitting ? 'Verifying...' : 'Unlock Workspace'}
					</button>

					<button
						type="button"
						onclick={useDefaultKey}
						class="text-xs text-zinc-400 hover:text-zinc-200 transition-colors py-1 underline underline-offset-4"
					>
						Use default local key (autodidact-local-key)
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
