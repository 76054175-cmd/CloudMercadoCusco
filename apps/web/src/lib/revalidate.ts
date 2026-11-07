import { cart } from './cart.svelte';
import { formatPEN } from './constants';
import { productImage } from './images';
import { pb } from './pb';
import type { Market, Product } from './types';

export type Change =
	| { kind: 'price'; name: string; from: number; to: number }
	| { kind: 'out_of_stock'; name: string }
	| { kind: 'gone'; name: string }
	| { kind: 'back'; name: string };

export function describe(c: Change): string {
	switch (c.kind) {
		case 'price':
			return `${c.name}: el precio cambió de ${formatPEN(c.from)} a ${formatPEN(c.to)}.`;
		case 'out_of_stock':
			return `${c.name} se agotó en el puesto.`;
		case 'gone':
			return `${c.name} ya no está publicado.`;
		case 'back':
			return `${c.name} volvió a estar disponible.`;
	}
}

const CHUNK = 40; // ids por consulta (límite práctico del largo del filtro)

/**
 * Compara el carrito con PocketBase: precio, stock y si el producto/puesto/mercado
 * sigue activo. Actualiza el carrito y devuelve la lista de cambios para avisar.
 */
export async function revalidateCart(): Promise<Change[]> {
	const ids = cart.items.map((i) => i.productId);
	if (ids.length === 0) return [];

	const fresh = new Map<string, Product>();
	for (let i = 0; i < ids.length; i += CHUNK) {
		const chunk = ids.slice(i, i + CHUNK);
		const filter = chunk.map((_, n) => `id = {:id${n}}`).join(' || ');
		const params = Object.fromEntries(chunk.map((id, n) => [`id${n}`, id]));
		const rows = await pb.collection('products').getFullList<Product>({
			filter: pb.filter(filter, params),
			expand: 'stall_id.market_id,catalog_item_id'
		});
		for (const r of rows) fresh.set(r.id, r);
	}

	const changes: Change[] = [];
	for (const item of [...cart.items]) {
		const p = fresh.get(item.productId);
		const stall = p?.expand?.stall_id;
		const market = stall?.expand?.market_id as Market | undefined;
		const wasAvailable = item.available !== false;

		// La regla de la API ya oculta productos de puestos inactivos; también se valida el mercado.
		if (!p || !stall?.active || !market?.active) {
			if (wasAvailable) changes.push({ kind: 'gone', name: item.name });
			cart.patch(item.productId, { available: false });
			continue;
		}

		if (p.price !== item.price) {
			changes.push({ kind: 'price', name: p.name, from: item.price, to: p.price });
		}
		if (!p.in_stock && wasAvailable) changes.push({ kind: 'out_of_stock', name: p.name });
		if (p.in_stock && !wasAvailable) changes.push({ kind: 'back', name: p.name });

		cart.patch(item.productId, {
			price: p.price,
			name: p.name,
			unitType: p.unit_type,
			image: productImage(p),
			stallName: stall.name,
			stallNumber: stall.stall_number,
			marketName: market.name,
			marketSlug: market.slug,
			available: p.in_stock
		});
	}
	return changes;
}
