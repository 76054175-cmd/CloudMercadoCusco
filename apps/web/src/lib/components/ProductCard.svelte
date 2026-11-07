<script lang="ts">
	import { resolve } from '$app/paths';
	import { cart } from '../cart.svelte';
	import { ratingLabel, toCartItem } from '../catalog';
	import { UNIT_TYPES, formatPEN } from '../constants';
	import { productImage } from '../images';
	import type { Market, Product } from '../types';
	import QtyStepper from './QtyStepper.svelte';

	interface Props {
		product: Product;
		market: Market;
		/** Cuántos otros puestos venden el mismo catalog_item. */
		otherStalls?: number;
		/** Mostrar el mercado (en la vista "otros puestos"). */
		showMarket?: boolean;
		/** Resaltar como el más barato. */
		cheapest?: boolean;
	}
	let { product, market, otherStalls = 0, showMarket = false, cheapest = false }: Props = $props();

	const stall = $derived(product.expand?.stall_id);
	const qty = $derived(cart.qtyOf(product.id));
</script>

<li
	class="card flex flex-col overflow-hidden border bg-surface-100-900 {cheapest
		? 'border-success-500'
		: 'border-surface-200-800'}"
>
	<div class="relative">
		<img
			src={productImage(product)}
			alt={product.name}
			loading="lazy"
			class="aspect-square w-full object-cover {product.in_stock ? '' : 'opacity-60 grayscale'}"
		/>
		{#if !product.in_stock}
			<span class="absolute top-2 left-2 badge preset-filled-error-500">Agotado</span>
		{:else if cheapest}
			<span class="absolute top-2 left-2 badge preset-filled-success-500">Mejor precio</span>
		{/if}
	</div>

	<div class="flex flex-1 flex-col gap-1 p-3">
		<h3 class="leading-tight font-semibold">{product.name}</h3>
		<p>
			<span class="text-lg font-bold">{formatPEN(product.price)}</span>
			<span class="text-sm opacity-70">/ {UNIT_TYPES[product.unit_type]}</span>
		</p>
		{#if stall}
			<p class="text-xs opacity-80">
				{#if showMarket}<span class="font-semibold">{market.name}</span> ·{/if}
				{stall.name}
				<span class="whitespace-nowrap" title="Calificación del puesto (escala 1 a 3)">
					· ⭐ {ratingLabel(stall.metrics_summary)}
				</span>
			</p>
		{/if}
		{#if otherStalls > 0 && product.catalog_item_id}
			<a
				class="anchor text-xs"
				href={resolve('/producto/[id]', { id: product.catalog_item_id })}
			>
				Ver en {otherStalls === 1 ? 'otro puesto' : `otros ${otherStalls} puestos`}
			</a>
		{/if}

		<div class="mt-auto pt-2">
			{#if qty > 0}
				<QtyStepper {qty} label={product.name} onchange={(n) => cart.setQty(product.id, n)} />
			{:else}
				<button
					type="button"
					class="btn btn-sm w-full preset-filled-primary-500"
					disabled={!product.in_stock}
					onclick={() => cart.add(toCartItem(product, market))}
				>
					{product.in_stock ? 'Agregar' : 'No disponible'}
				</button>
			{/if}
		</div>
	</div>
</li>
