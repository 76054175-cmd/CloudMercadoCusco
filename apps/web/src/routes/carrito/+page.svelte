<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { cart, isAvailable, money } from '#lib/cart.svelte.js';
	import QtyStepper from '#lib/components/QtyStepper.svelte';
	import { UNIT_TYPES, formatPEN } from '#lib/constants.js';
	import { describe, revalidateCart, type Change } from '#lib/revalidate.js';
	import { settingNumber } from '#lib/settings.js';

	let status = $state<'checking' | 'ok' | 'error'>('checking');
	let changes = $state<Change[]>([]);
	let minOrder = $state(10);

	async function check() {
		status = 'checking';
		try {
			changes = await revalidateCart();
			status = 'ok';
		} catch {
			status = 'error';
		}
	}

	onMount(() => {
		settingNumber('min_order_amount', 10).then((v) => (minOrder = v));
		check();
	});

	const marketsBelowMin = $derived(cart.byMarket.filter((m) => m.subtotal < minOrder));
</script>

<svelte:head>
	<title>Mi carrito · QhatuDigital</title>
</svelte:head>

<section class="mx-auto max-w-4xl px-4 py-6">
	<div class="flex flex-wrap items-baseline justify-between gap-2">
		<h1 class="h2">🧺 Mi carrito</h1>
		{#if cart.items.length > 0}
			<button
				type="button"
				class="btn btn-sm preset-tonal-error"
				onclick={() => confirm('¿Vaciar todo el carrito?') && cart.clear()}
			>
				Vaciar carrito
			</button>
		{/if}
	</div>

	{#if status === 'checking' && cart.items.length > 0}
		<p class="mt-3 text-sm opacity-70" aria-live="polite">Revisando precios y stock…</p>
	{:else if status === 'error'}
		<div class="mt-3 card preset-tonal-warning p-3 text-sm" role="alert">
			No pudimos verificar precios y stock (¿sin conexión?). Los montos pueden no estar al día.
			<button type="button" class="anchor ml-1" onclick={check}>Reintentar</button>
		</div>
	{/if}

	{#if changes.length > 0}
		<div class="mt-3 card preset-tonal-warning p-4" role="alert">
			<p class="font-semibold">Algunos productos cambiaron desde que los agregaste:</p>
			<ul class="mt-2 list-disc space-y-1 pl-5 text-sm">
				{#each changes as c, i (i)}
					<li>{describe(c)}</li>
				{/each}
			</ul>
			<button type="button" class="btn btn-sm mt-3 preset-tonal" onclick={() => (changes = [])}>
				Entendido
			</button>
		</div>
	{/if}

	{#if cart.items.length === 0}
		<div class="mt-12 text-center">
			<p class="text-lg opacity-80">Tu carrito está vacío.</p>
			<a href={resolve('/')} class="btn mt-4 preset-filled-primary-500">Elegir un mercado</a>
		</div>
	{:else}
		{#each cart.byMarket as m (m.marketId)}
			{@const missing = money(minOrder - m.subtotal)}
			<article class="mt-6 card border border-surface-200-800 bg-surface-100-900">
				<header
					class="flex flex-wrap items-baseline justify-between gap-2 border-b border-surface-200-800 p-4"
				>
					<h2 class="h4">
						<a class="hover:underline" href={resolve('/mercado/[slug]', { slug: m.marketSlug })}>
							Mercado {m.marketName}
						</a>
					</h2>
					<span class="font-semibold">{formatPEN(m.subtotal)}</span>
				</header>

				{#each m.stalls as s (s.stallId)}
					<div class="border-b border-surface-200-800 px-4 py-3 last:border-b-0">
						<p class="text-sm font-semibold">
							{s.stallName} <span class="font-normal opacity-60">· {s.stallNumber}</span>
						</p>
						<ul class="mt-2 space-y-3">
							{#each s.items as item (item.productId)}
								{@const ok = isAvailable(item)}
								<li class="flex items-center gap-3">
									<img
										src={item.image}
										alt=""
										class="size-14 shrink-0 rounded object-cover {ok ? '' : 'opacity-50 grayscale'}"
									/>
									<div class="min-w-0 flex-1">
										<p class="truncate font-medium {ok ? '' : 'line-through opacity-60'}">
											{item.name}
										</p>
										<p class="text-xs opacity-70">
											{formatPEN(item.price)} / {UNIT_TYPES[item.unitType]}
										</p>
										{#if !ok}
											<span class="badge preset-tonal-error text-xs">No disponible · no se cobra</span>
										{/if}
									</div>
									{#if ok}
										<QtyStepper
											qty={item.qty}
											label={item.name}
											onchange={(n) => cart.setQty(item.productId, n)}
										/>
										<span class="w-20 text-right text-sm font-semibold tabular-nums">
											{formatPEN(money(item.price * item.qty))}
										</span>
									{:else}
										<button
											type="button"
											class="btn btn-sm preset-tonal"
											onclick={() => cart.remove(item.productId)}
										>
											Quitar
										</button>
									{/if}
								</li>
							{/each}
						</ul>
						<p class="mt-2 text-right text-xs opacity-70">Subtotal puesto: {formatPEN(s.subtotal)}</p>
					</div>
				{/each}

				<footer class="space-y-2 p-4">
					{#if m.subtotal < minOrder}
						<p class="card preset-tonal-warning p-2 text-sm">
							El pedido mínimo por mercado es {formatPEN(minOrder)}. Te faltan
							<strong>{formatPEN(missing)}</strong> en {m.marketName}.
						</p>
					{/if}
					<div class="flex flex-wrap items-center justify-between gap-2">
						<a
							class="anchor text-sm"
							href={resolve('/mercado/[slug]', { slug: m.marketSlug })}
						>
							+ Agregar más de {m.marketName}
						</a>
						<button
							type="button"
							class="btn preset-filled-primary-500"
							disabled
							title="El pago se habilita en la siguiente versión"
						>
							Pagar {m.marketName} (pronto)
						</button>
					</div>
				</footer>
			</article>
		{/each}

		<div class="mt-6 card preset-tonal p-4">
			<div class="flex items-baseline justify-between">
				<span class="text-lg font-semibold">Total general</span>
				<span class="text-2xl font-bold">{formatPEN(cart.total)}</span>
			</div>
			<p class="mt-1 text-sm opacity-70">
				{cart.byMarket.length}
				{cart.byMarket.length === 1 ? 'mercado' : 'mercados'} · sin costo de envío. Cada mercado
				se paga y se recoge por separado, en su puerta principal.
			</p>
			{#if marketsBelowMin.length > 0}
				<p class="mt-1 text-sm opacity-70">
					{marketsBelowMin.length === 1 ? 'Un mercado no llega' : `${marketsBelowMin.length} mercados no llegan`}
					al mínimo de {formatPEN(minOrder)}.
				</p>
			{/if}
		</div>
	{/if}
</section>
