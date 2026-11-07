import { pb } from './pb';
import { SECTIONS } from './constants';
import type { Market, Product } from './types';

/** Placeholder SVG (data URI) con el ícono de la sección. */
export function placeholder(icon: string): string {
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect width="100" height="100" fill="#efe7da"/><text x="50" y="62" font-size="42" text-anchor="middle">${icon}</text></svg>`;
	return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
}

/**
 * Imagen de producto con fallback:
 * 1) foto subida al producto  2) imagen del catalog_item  3) placeholder por sección.
 */
export function productImage(p: Product): string {
	if (p.image) return pb.files.getURL(p, p.image, { thumb: '400x400' });
	const ci = p.expand?.catalog_item_id;
	if (ci?.default_image) return pb.files.getURL(ci, ci.default_image, { thumb: '400x400' });
	const section = ci?.section ?? p.expand?.stall_id?.section ?? 'otros';
	return placeholder(SECTIONS[section].icon);
}

export function marketIcon(m: Market): string | null {
	return m.bubble_icon ? pb.files.getURL(m, m.bubble_icon, { thumb: '120x120' }) : null;
}
