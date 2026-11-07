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

	const [stalls, products] = await Promise.all([
		pb.collection('stalls').getFullList<Stall>({
			filter: pb.filter('market_id = {:id} && active = true', { id: market.id }),
			sort: 'name'
		}),
		pb.collection('products').getFullList<Product>({
			filter: pb.filter('stall_id.market_id = {:id} && stall_id.active = true', { id: market.id }),
			expand: 'stall_id,catalog_item_id',
			sort: 'name'
		})
	]);

	return { market, stalls, products };
};
