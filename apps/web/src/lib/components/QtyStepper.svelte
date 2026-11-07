<script lang="ts">
	import { MAX_QTY } from '../cart.svelte';

	interface Props {
		qty: number;
		label: string;
		onchange: (qty: number) => void;
	}
	let { qty, label, onchange }: Props = $props();
</script>

<div class="flex items-center gap-1" role="group" aria-label="Cantidad de {label}">
	<button
		type="button"
		class="btn-icon btn-icon-sm preset-tonal"
		aria-label={qty === 1 ? `Quitar ${label}` : `Restar uno de ${label}`}
		onclick={() => onchange(qty - 1)}
	>
		{qty === 1 ? '🗑' : '−'}
	</button>
	<span class="min-w-8 text-center font-semibold tabular-nums" aria-live="polite">{qty}</span>
	<button
		type="button"
		class="btn-icon btn-icon-sm preset-filled-primary-500"
		aria-label="Sumar uno de {label}"
		disabled={qty >= MAX_QTY}
		onclick={() => onchange(qty + 1)}
	>
		+
	</button>
</div>
