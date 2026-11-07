import type { CartItem } from './cart.svelte';
import { productImage } from './images';
import type { Market, MetricsSummary, Product, Stall } from './types';

/** Minúsculas y sin tildes: "Plátano" y "platano" coinciden. */
export const normalize = (s: string) =>
	s.normalize('NFD').replace(/\p{Diacritic}/gu, '').toLowerCase().trim();

/** Convierte un producto (con expand stall_id) en ítem de carrito. */
export function toCartItem(p: Product, market: Market): Omit<CartItem, 'qty'> {
	const stall = p.expand?.stall_id as Stall;
	return {
		productId: p.id,
		name: p.name,
		price: p.price,
		unitType: p.unit_type,
		image: productImage(p),
		stallId: p.stall_id,
		stallName: stall?.name ?? '',
		stallNumber: stall?.stall_number ?? '',
		marketId: market.id,
		marketName: market.name,
		marketSlug: market.slug
	};
}

/** Calificación minimalista (escala 1–3). Sin votos → "Nuevo". */
export function ratingLabel(m: MetricsSummary | null | undefined): string {
	if (!m || !m.total_votes) return 'Nuevo';
	return `${m.overall.toFixed(1)}/3`;
}
