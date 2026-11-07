<script lang="ts">
	import './layout.css';
	import favicon from '#lib/assets/favicon.svg';
	import { resolve } from '$app/paths';
	import { cart } from '#lib/cart.svelte.js';
	import { formatPEN } from '#lib/constants.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="flex min-h-dvh flex-col bg-surface-50-950">
	<header class="border-b border-surface-200-800 bg-surface-100-900/80 backdrop-blur">
		<div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
			<a href={resolve('/')} class="flex items-center gap-2 font-bold">
				<span class="text-2xl" aria-hidden="true">🧺</span>
				<span class="text-lg">Qhatu<span class="text-primary-600-400">Digital</span></span>
			</a>
			<a
				href={resolve('/carrito')}
				class="btn btn-sm preset-tonal-primary"
				aria-label="Carrito: {cart.count} productos, {formatPEN(cart.total)}"
			>
				<span aria-hidden="true">🧺</span>
				{#if cart.count > 0}
					<span class="badge-icon preset-filled-primary-500 tabular-nums">{cart.count}</span>
					<span class="hidden sm:inline">{formatPEN(cart.total)}</span>
				{:else}
					<span>Carrito</span>
				{/if}
			</a>
		</div>
	</header>

	<main class="flex-1">
		{@render children()}
	</main>

	<footer class="py-4 text-center text-xs opacity-60">
		QhatuDigital Cusco · Compra en el mercado y recoge en la puerta principal
	</footer>
</div>
