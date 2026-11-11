<script lang="ts">
	import { Play, Pause, RotateCcw, Clock } from '@lucide/svelte';
	import { onDestroy } from 'svelte';

	let mode = $state<'pomodoro' | 'stopwatch'>('pomodoro');
	let isRunning = $state(false);
	let secondsLeft = $state(25 * 60);
	let elapsedSeconds = $state(0);
	let timerInterval: any = null;

	function toggleStart() {
		if (isRunning) {
			pause();
		} else {
			start();
		}
	}

	function start() {
		isRunning = true;
		clearInterval(timerInterval);
		timerInterval = setInterval(() => {
			if (mode === 'pomodoro') {
				if (secondsLeft > 0) {
					secondsLeft--;
				} else {
					pause();
				}
			} else {
				elapsedSeconds++;
			}
		}, 1000);
	}

	function pause() {
		isRunning = false;
		clearInterval(timerInterval);
	}

	function reset() {
		pause();
		if (mode === 'pomodoro') {
			secondsLeft = 25 * 60;
		} else {
			elapsedSeconds = 0;
		}
	}

	function switchMode(newMode: 'pomodoro' | 'stopwatch') {
		pause();
		mode = newMode;
		if (newMode === 'pomodoro') {
			secondsLeft = 25 * 60;
		} else {
			elapsedSeconds = 0;
		}
	}

	function formatTime(totalSec: number): string {
		const m = Math.floor(totalSec / 60);
		const s = totalSec % 60;
		return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
	}

	onDestroy(() => {
		clearInterval(timerInterval);
	});
</script>

<div class="inline-flex items-center gap-2 rounded-xl border border-zinc-800 bg-zinc-900/90 px-3 py-1.5 shadow-xs">
	<button
		type="button"
		onclick={() => switchMode(mode === 'pomodoro' ? 'stopwatch' : 'pomodoro')}
		class="text-xs font-medium text-zinc-400 hover:text-indigo-400 transition-colors uppercase tracking-wider flex items-center gap-1"
		title="Click to toggle Pomodoro / Stopwatch"
	>
		<Clock class="h-3.5 w-3.5" />
		<span>{mode === 'pomodoro' ? 'Pomo' : 'Stopwatch'}</span>
	</button>

	<span class="font-mono text-sm font-semibold tracking-tight {isRunning ? 'text-indigo-400' : 'text-zinc-200'}">
		{formatTime(mode === 'pomodoro' ? secondsLeft : elapsedSeconds)}
	</span>

	<div class="flex items-center gap-1">
		<button
			type="button"
			onclick={toggleStart}
			class="flex h-7 w-7 items-center justify-center rounded-lg hover:bg-zinc-800 text-zinc-300 hover:text-white transition-colors"
			title={isRunning ? 'Pause' : 'Start'}
		>
			{#if isRunning}
				<Pause class="h-3.5 w-3.5 fill-current" />
			{:else}
				<Play class="h-3.5 w-3.5 fill-current" />
			{/if}
		</button>
		<button
			type="button"
			onclick={reset}
			class="flex h-7 w-7 items-center justify-center rounded-lg hover:bg-zinc-800 text-zinc-400 hover:text-white transition-colors"
			title="Reset"
		>
			<RotateCcw class="h-3.5 w-3.5" />
		</button>
	</div>
</div>
