import { error } from '@sveltejs/kit';
import { pb } from '#lib/pb.js';
import type { Market, Product, Stall } from '#lib/types.js';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	let market: Market;
	try {
		market = await pb
			.collection('markets')
			.getFirstListItem<Market>(pb.filter('slug = {:slug}', { slug: params.slug }));
	} catch {
		error(404, 'Mercado no encontrado');
	}

	const [stalls, products, offers] = await Promise.all([
		pb.collection('stalls').getFullList<Stall>({
			filter: pb.filter('market_id = {:id} && active = true', { id: market.id }),
			sort: 'name'
		}),
		pb.collection('products').getFullList<Product>({
			filter: pb.filter('stall_id.market_id = {:id} && stall_id.active = true', { id: market.id }),
			expand: 'stall_id,catalog_item_id',
			sort: '-in_stock,name'
		}),
		// Para "Ver en otros puestos": qué puestos activos (de cualquier mercado) ofrecen cada catalog_item.
		pb.collection('products').getFullList<Pick<Product, 'catalog_item_id' | 'stall_id'>>({
			filter: 'stall_id.active = true && stall_id.market_id.active = true && catalog_item_id != ""',
			fields: 'catalog_item_id,stall_id'
		})
	]);

	const stallsByItem = new Map<string, Set<string>>();
	for (const o of offers) {
		const set = stallsByItem.get(o.catalog_item_id) ?? new Set<string>();
		set.add(o.stall_id);
		stallsByItem.set(o.catalog_item_id, set);
	}
	/** Puestos distintos a stallId que venden el mismo catalog_item. */
	const otherStalls = (catalogItemId: string, stallId: string) => {
		const set = stallsByItem.get(catalogItemId);
		return set ? set.size - (set.has(stallId) ? 1 : 0) : 0;
	};

	return { market, stalls, products, otherStalls };
};
