import { pb } from '#lib/pb.js';
import type { Market } from '#lib/types.js';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	const markets = await pb.collection('markets').getFullList<Market>({ sort: 'created' });
	return { markets };
};
