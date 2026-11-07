<script lang="ts">
	import { resolve } from '$app/paths';
	import ProductCard from '#lib/components/ProductCard.svelte';
	import { SECTIONS, UNIT_TYPES, formatPEN } from '#lib/constants.js';
	import type { Market } from '#lib/types.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	// "Mejor precio" solo compara la misma unidad (kg con kg) y si hay al menos 2 en stock.
	const cheapestIds = $derived.by(() => {
		const ids = new Set<string>();
		const inStock = data.offers.filter((p) => p.in_stock);
		for (const group of Map.groupBy(inStock, (p) => p.unit_type).values()) {
			if (group.length < 2) continue;
			const min = Math.min(...group.map((p) => p.price));
			for (const p of group) if (p.price === min) ids.add(p.id);
		}
		return ids;
	});

	const stallCount = $derived(new Set(data.offers.map((p) => p.stall_id)).size);
	const units = $derived([...new Set(data.offers.map((p) => p.unit_type))]);
	const range = $derived.by(() => {
		const prices = data.offers.filter((p) => p.in_stock).map((p) => p.price);
		return prices.length ? { min: Math.min(...prices), max: Math.max(...prices) } : null;
	});
</script>

<svelte:head>
	<title>{data.item.name} en todos los puestos · QhatuDigital</title>
</svelte:head>

<section class="mx-auto max-w-6xl px-4 py-6">
	<button type="button" class="anchor text-sm" onclick={() => history.back()}>← Volver</button>

	<header class="mt-3">
		<h1 class="h2">{SECTIONS[data.item.section].icon} {data.item.name}</h1>
		<p class="mt-1 opacity-80">
			{stallCount}
			{stallCount === 1 ? 'puesto lo ofrece' : 'puestos lo ofrecen'}, ordenados por precio.
			{#if range && units.length === 1 && range.min !== range.max}
				Desde <strong>{formatPEN(range.min)}</strong> hasta {formatPEN(range.max)} / {UNIT_TYPES[units[0]]}.
			{/if}
		</p>
		<p class="mt-1 text-sm opacity-70">
			Tu carrito acepta productos de varios mercados a la vez: cada mercado se paga y recoge por separado.
		</p>
	</header>

	{#if data.offers.length === 0}
		<p class="mt-10 text-center opacity-70">Ningún puesto activo ofrece este producto por ahora.</p>
	{:else}
		<ul class="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
			{#each data.offers as product (product.id)}
				{@const market = product.expand?.stall_id?.expand?.market_id as Market}
				<ProductCard
					{product}
					{market}
					showMarket
					cheapest={cheapestIds.has(product.id)}
				/>
			{/each}
		</ul>
		<p class="mt-6 text-sm">
			<a class="anchor" href={resolve('/carrito')}>Ir al carrito →</a>
		</p>
	{/if}
</section>
