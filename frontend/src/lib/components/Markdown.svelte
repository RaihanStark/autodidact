<script lang="ts">
	import { marked } from 'marked';

	let { content = '' }: { content?: string } = $props();

	let htmlContent = $derived.by(() => {
		if (!content) return '';
		try {
			return marked.parse(content, { breaks: true, gfm: true }) as string;
		} catch (e) {
			return `<p>${content}</p>`;
		}
	});
</script>

<div class="prose prose-invert prose-zinc max-w-none prose-headings:font-semibold prose-a:text-indigo-400 prose-a:no-underline hover:prose-a:underline prose-pre:bg-zinc-950 prose-pre:border prose-pre:border-zinc-800">
	{@html htmlContent}
</div>
