<script lang="ts">
	import { resolve } from '$app/paths';
	import { marketIcon } from '#lib/images.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	// Tonos y tamaños que se repiten para dar un aspecto orgánico a las burbujas.
	const tones = [
		'from-primary-400 to-primary-600',
		'from-secondary-400 to-secondary-600',
		'from-tertiary-400 to-tertiary-600',
		'from-success-400 to-success-600',
		'from-warning-400 to-warning-600'
	];
	const sizes = ['size-36 md:size-44', 'size-32 md:size-40', 'size-40 md:size-48'];
</script>

<section class="mx-auto flex max-w-5xl flex-col items-center px-4 py-10 text-center md:py-16">
	<h1 class="h2 md:h1">¿A qué mercado vas hoy?</h1>
	<p class="mt-3 max-w-xl opacity-75">
		Elige un mercado, arma tu canasta con productos de varios puestos y recoge todo junto en la
		puerta principal.
	</p>

	{#if data.markets.length === 0}
		<p class="mt-12 opacity-70">Aún no hay mercados habilitados.</p>
	{:else}
		<ul class="mt-12 flex flex-wrap items-center justify-center gap-6 md:gap-10" aria-label="Mercados">
			{#each data.markets as market, i (market.id)}
				{@const icon = marketIcon(market)}
				<li
					class="bubble-float"
					style:--float-delay="{(i * 0.7) % 3}s"
					style:--float-duration="{5 + (i % 3)}s"
				>
					<a
						href={resolve('/mercado/[slug]', { slug: market.slug })}
						class="flex {sizes[i % sizes.length]} flex-col items-center justify-center rounded-full bg-linear-to-br {tones[
							i % tones.length
						]} p-4 text-white shadow-xl ring-4 ring-white/40 transition hover:scale-110 hover:shadow-2xl focus-visible:scale-110 focus-visible:ring-8 focus-visible:outline-none"
					>
						{#if icon}
							<img src={icon} alt="" class="mb-1 size-12 rounded-full object-cover" />
						{:else}
							<span class="mb-1 text-3xl" aria-hidden="true">🏪</span>
						{/if}
						<span class="text-lg leading-tight font-bold drop-shadow md:text-xl">{market.name}</span>
						<span class="mt-1 text-xs opacity-90">{market.open_time}–{market.close_time}</span>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</section>
