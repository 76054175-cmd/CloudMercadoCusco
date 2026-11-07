import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

// En desarrollo Vite reenvía la API y el panel de PocketBase al core (Go).
// En producción la web está embebida en el mismo binario: mismo origen, sin CORS.
const coreTarget = process.env.PB_PROXY_TARGET ?? 'http://localhost:8090';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// SPA estática: preparada para Capacitor en la v2.
			adapter: adapter({ fallback: 'index.html' })
		})
	],
	server: {
		proxy: {
			'/api': { target: coreTarget, changeOrigin: true },
			'/_': { target: coreTarget, changeOrigin: true }
		}
	}
});
