import { error } from '@sveltejs/kit';
import { pb } from '#lib/pb.js';
import type { CatalogItem, Product } from '#lib/types.js';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	let item: CatalogItem;
	try {
		item = await pb.collection('catalog_items').getOne<CatalogItem>(params.id);
	} catch {
		error(404, 'Producto no encontrado');
	}

	const offers = await pb.collection('products').getFullList<Product>({
		filter: pb.filter(
			'catalog_item_id = {:id} && stall_id.active = true && stall_id.market_id.active = true',
			{ id: item.id }
		),
		expand: 'stall_id.market_id,catalog_item_id',
		sort: '-in_stock,price'
	});

	return { item, offers };
};
