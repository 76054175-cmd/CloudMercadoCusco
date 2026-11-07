<script lang="ts">
	import { resolve } from '$app/paths';
	import { SECTIONS, UNIT_TYPES, formatPEN } from '#lib/constants.js';
	import { productImage } from '#lib/images.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const productsByStall = $derived(
		Map.groupBy(data.products, (p) => p.stall_id)
	);
</script>

<svelte:head>
	<title>{data.market.name} · QhatuDigital</title>
</svelte:head>

<section class="mx-auto max-w-6xl px-4 py-6">
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
	{/if}

	{#each data.stalls as stall (stall.id)}
		{@const items = productsByStall.get(stall.id) ?? []}
		<article class="mt-8">
			<header class="flex flex-wrap items-baseline justify-between gap-2 border-b border-surface-200-800 pb-2">
				<div>
					<h2 class="h4">{SECTIONS[stall.section].icon} {stall.name}</h2>
					<p class="text-xs opacity-70">{stall.stall_number} · {SECTIONS[stall.section].label}</p>
				</div>
				{#if stall.metrics_summary}
					<span class="text-sm" title="Calificación promedio (1 a 3)">
						⭐ {stall.metrics_summary.overall.toFixed(1)}/3
						<span class="opacity-60">({stall.metrics_summary.total_votes} votos)</span>
					</span>
				{/if}
			</header>

			{#if items.length === 0}
				<p class="mt-3 text-sm opacity-60">Sin productos publicados.</p>
			{:else}
				<ul class="mt-4 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
					{#each items as product (product.id)}
						<li class="card overflow-hidden border border-surface-200-800 bg-surface-100-900">
							<img
								src={productImage(product)}
								alt={product.name}
								loading="lazy"
								class="aspect-square w-full object-cover {product.in_stock ? '' : 'grayscale opacity-60'}"
							/>
							<div class="p-3">
								<h3 class="leading-tight font-semibold">{product.name}</h3>
								<p class="mt-1 text-sm">
									<span class="font-bold">{formatPEN(product.price)}</span>
									<span class="opacity-70">/ {UNIT_TYPES[product.unit_type]}</span>
								</p>
								{#if !product.in_stock}
									<span class="mt-2 badge preset-tonal-error">Agotado</span>
								{/if}
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		</article>
	{/each}
</section>
