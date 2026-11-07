<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
</script>

<section class="mx-auto max-w-md px-4 py-16 text-center">
	<p class="text-5xl" aria-hidden="true">🧺</p>
	<h1 class="mt-4 h3">{page.status === 404 ? 'No encontramos esa página' : 'Algo salió mal'}</h1>
	<p class="mt-2 opacity-70">{page.error?.message}</p>
	<a href={resolve('/')} class="mt-6 btn preset-filled-primary-500">Volver a los mercados</a>
</section>
