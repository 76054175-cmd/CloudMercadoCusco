<script lang="ts">
	import { resolve } from '$app/paths';
	import { cart } from '#lib/cart.svelte.js';
	import { normalize, ratingLabel } from '#lib/catalog.js';
	import ProductCard from '#lib/components/ProductCard.svelte';
	import { SECTIONS, formatPEN, type Section } from '#lib/constants.js';
	import type { Product } from '#lib/types.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let query = $state('');
	let section = $state<Section | 'all'>('all');
	let stallId = $state('all');

	const sectionOf = (p: Product): Section =>
		p.expand?.catalog_item_id?.section ?? p.expand?.stall_id?.section ?? 'otros';

	// Texto donde busca el buscador: producto, catálogo y puesto.
	const haystack = $derived(
		new Map(
			data.products.map((p) => [
				p.id,
				normalize(`${p.name} ${p.expand?.catalog_item_id?.name ?? ''} ${p.expand?.stall_id?.name ?? ''}`)
			])
		)
	);

	const sectionCounts = $derived(
		Map.groupBy(data.products, sectionOf)
	);

	const filtered = $derived.by(() => {
		const q = normalize(query);
		return data.products.filter(
			(p) =>
				(section === 'all' || sectionOf(p) === section) &&
				(stallId === 'all' || p.stall_id === stallId) &&
				(!q || haystack.get(p.id)!.includes(q))
		);
	});

	const byStall = $derived(Map.groupBy(filtered, (p) => p.stall_id));
	const visibleStalls = $derived(data.stalls.filter((s) => byStall.has(s.id)));
	const hasFilters = $derived(query !== '' || section !== 'all' || stallId !== 'all');
	const inThisMarket = $derived(cart.byMarket.find((m) => m.marketId === data.market.id));

	function resetFilters() {
		query = '';
		section = 'all';
		stallId = 'all';
	}
</script>

<svelte:head>
	<title>{data.market.name} · QhatuDigital</title>
</svelte:head>

<section class="mx-auto max-w-6xl px-4 py-6 {inThisMarket ? 'pb-24' : ''}">
	<a href={resolve('/')} class="anchor text-sm">← Cambiar de mercado</a>

	<div class="mt-3 card preset-tonal p-5">
		<h1 class="h2">Mercado {data.market.name}</h1>
		<p class="mt-2 text-sm opacity-80">
			📍 Recojo: {data.market.hub_location_desc} · 🕒 {data.market.open_time}–{data.market.close_time}
		</p>
		<div class="mt-3 flex flex-wrap gap-2">
			<span class="badge preset-filled-primary-500">Recojo en puerta principal</span>
			{#if data.market.has_own_delivery}
				<span class="badge preset-tonal-success">Delivery del mercado · {formatPEN(data.market.delivery_fee)}</span>
			{/if}
			{#if data.market.rappi_enabled}
				<span class="badge preset-tonal-tertiary">Rappi</span>
			{/if}
		</div>
	</div>

	{#if data.stalls.length === 0}
		<p class="mt-10 text-center opacity-70">
			Este mercado todavía no tiene puestos habilitados. ¡Pronto!
		</p>
	{:else}
		<!-- Barra de filtros -->
		<div
			class="sticky top-0 z-10 -mx-4 mt-6 space-y-3 border-b border-surface-200-800 bg-surface-50-950/95 px-4 py-3 backdrop-blur"
		>
			<div class="flex flex-col gap-2 sm:flex-row">
				<label class="flex-1">
					<span class="sr-only">Buscar producto</span>
					<input
						class="input"
						type="search"
						placeholder="Buscar: papa, queso, pollo…"
						bind:value={query}
						autocomplete="off"
					/>
				</label>
				<label class="sm:w-64">
					<span class="sr-only">Filtrar por puesto</span>
					<select class="select" bind:value={stallId}>
						<option value="all">Todos los puestos ({data.stalls.length})</option>
						{#each data.stalls as s (s.id)}
							<option value={s.id}>{s.name}</option>
						{/each}
					</select>
				</label>
			</div>

			<div class="flex gap-2 overflow-x-auto pb-1" role="group" aria-label="Secciones">
				<button
					type="button"
					class="chip shrink-0 {section === 'all' ? 'preset-filled-primary-500' : 'preset-tonal'}"
					aria-pressed={section === 'all'}
					onclick={() => (section = 'all')}
				>
					Todo ({data.products.length})
				</button>
				{#each Object.entries(SECTIONS) as [key, s] (key)}
					{@const n = sectionCounts.get(key as Section)?.length ?? 0}
					{#if n > 0}
						<button
							type="button"
							class="chip shrink-0 {section === key ? 'preset-filled-primary-500' : 'preset-tonal'}"
							aria-pressed={section === key}
							onclick={() => (section = section === key ? 'all' : (key as Section))}
						>
							{s.icon} {s.label} ({n})
						</button>
					{/if}
				{/each}
			</div>
		</div>

		<p class="mt-3 text-sm opacity-70" aria-live="polite">
			{filtered.length} de {data.products.length} productos
			{#if hasFilters}
				· <button type="button" class="anchor" onclick={resetFilters}>Quitar filtros</button>
			{/if}
		</p>

		{#if filtered.length === 0}
			<div class="mt-10 text-center">
				<p class="opacity-70">No encontramos productos con esos filtros.</p>
				<button type="button" class="btn mt-3 preset-tonal" onclick={resetFilters}>Ver todo</button>
			</div>
		{/if}

		{#each visibleStalls as stall (stall.id)}
			<article class="mt-8">
				<header
					class="flex flex-wrap items-baseline justify-between gap-2 border-b border-surface-200-800 pb-2"
				>
					<div>
						<h2 class="h4">{SECTIONS[stall.section].icon} {stall.name}</h2>
						<p class="text-xs opacity-70">{stall.stall_number} · {SECTIONS[stall.section].label}</p>
					</div>
					<span class="text-sm" title="Calificación del puesto (escala 1 a 3)">
						⭐ {ratingLabel(stall.metrics_summary)}
						{#if stall.metrics_summary?.total_votes}
							<span class="opacity-60">({stall.metrics_summary.total_votes} votos)</span>
						{/if}
					</span>
				</header>

				<ul class="mt-4 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
					{#each byStall.get(stall.id) ?? [] as product (product.id)}
						<ProductCard
							{product}
							market={data.market}
							otherStalls={data.otherStalls(product.catalog_item_id, product.stall_id)}
						/>
					{/each}
				</ul>
			</article>
		{/each}
	{/if}
</section>

{#if inThisMarket}
	<a
		href={resolve('/carrito')}
		class="btn fixed right-4 bottom-4 z-20 preset-filled-primary-500 shadow-xl"
	>
		🧺 {inThisMarket.count} en {data.market.name} · {formatPEN(inThisMarket.subtotal)} — Ver carrito
	</a>
{/if}
