// Carrito global, independiente del mercado. Solo $state/$derived (Svelte 5 runes)
// y persistencia manual en localStorage en cada cambio.
import type { UnitType } from './constants';

const STORAGE_KEY = 'qhatu.cart.v1';
export const MAX_QTY = 99;

export interface CartItem {
	productId: string;
	name: string;
	price: number; // precio vigente en el puesto (se revalida al abrir el carrito)
	unitType: UnitType;
	image: string;
	qty: number;
	stallId: string;
	stallName: string;
	stallNumber: string;
	marketId: string;
	marketName: string;
	marketSlug: string;
	/** false si al revalidar el producto está agotado o ya no se publica. No suma al total. */
	available?: boolean;
}

export const isAvailable = (i: CartItem) => i.available !== false;

export interface StallGroup {
	stallId: string;
	stallName: string;
	stallNumber: string;
	items: CartItem[];
	subtotal: number;
}

export interface MarketGroup {
	marketId: string;
	marketName: string;
	marketSlug: string;
	stalls: StallGroup[];
	subtotal: number;
	count: number;
}

/** Redondeo a céntimos para evitar 0.1 + 0.2. */
export const money = (n: number) => Math.round(n * 100) / 100;

function load(): CartItem[] {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		const parsed: unknown = raw ? JSON.parse(raw) : [];
		return Array.isArray(parsed) ? parsed.filter(isItem) : [];
	} catch {
		return [];
	}
}

function isItem(x: unknown): x is CartItem {
	const i = x as CartItem;
	return (
		typeof i?.productId === 'string' &&
		typeof i.price === 'number' &&
		Number.isInteger(i.qty) &&
		i.qty > 0 &&
		typeof i.marketId === 'string' &&
		typeof i.stallId === 'string'
	);
}

class Cart {
	items = $state<CartItem[]>(load());

	count = $derived(this.items.reduce((n, i) => n + i.qty, 0));
	total = $derived(
		money(this.items.filter(isAvailable).reduce((s, i) => s + i.price * i.qty, 0))
	);

	/** Agrupado Mercado → Puesto, en el orden en que se agregaron. */
	byMarket = $derived.by(() => {
		const markets = new Map<string, MarketGroup>();
		for (const item of this.items) {
			let m = markets.get(item.marketId);
			if (!m) {
				m = {
					marketId: item.marketId,
					marketName: item.marketName,
					marketSlug: item.marketSlug,
					stalls: [],
					subtotal: 0,
					count: 0
				};
				markets.set(item.marketId, m);
			}
			let s = m.stalls.find((g) => g.stallId === item.stallId);
			if (!s) {
				s = {
					stallId: item.stallId,
					stallName: item.stallName,
					stallNumber: item.stallNumber,
					items: [],
					subtotal: 0
				};
				m.stalls.push(s);
			}
			s.items.push(item);
			if (!isAvailable(item)) continue;
			s.subtotal = money(s.subtotal + item.price * item.qty);
			m.subtotal = money(m.subtotal + item.price * item.qty);
			m.count += item.qty;
		}
		return [...markets.values()];
	});

	constructor() {
		// Mantener sincronizadas varias pestañas abiertas.
		window.addEventListener('storage', (e) => {
			if (e.key === STORAGE_KEY) this.items = load();
		});
	}

	qtyOf(productId: string): number {
		return this.items.find((i) => i.productId === productId)?.qty ?? 0;
	}

	add(item: Omit<CartItem, 'qty'>, qty = 1) {
		const existing = this.items.find((i) => i.productId === item.productId);
		if (existing) existing.qty = Math.min(MAX_QTY, existing.qty + qty);
		else this.items.push({ ...item, qty: Math.min(MAX_QTY, qty) });
		this.save();
	}

	setQty(productId: string, qty: number) {
		if (qty <= 0) return this.remove(productId);
		const item = this.items.find((i) => i.productId === productId);
		if (!item) return;
		item.qty = Math.min(MAX_QTY, Math.floor(qty));
		this.save();
	}

	/** Actualiza datos de un ítem (p. ej. precio revalidado) sin tocar la cantidad. */
	patch(productId: string, data: Partial<Omit<CartItem, 'productId' | 'qty'>>) {
		const item = this.items.find((i) => i.productId === productId);
		if (!item) return;
		Object.assign(item, data);
		this.save();
	}

	remove(productId: string) {
		this.items = this.items.filter((i) => i.productId !== productId);
		this.save();
	}

	clearMarket(marketId: string) {
		this.items = this.items.filter((i) => i.marketId !== marketId);
		this.save();
	}

	clear() {
		this.items = [];
		this.save();
	}

	private save() {
		try {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(this.items));
		} catch {
			// Modo privado o almacenamiento lleno: el carrito sigue en memoria.
		}
	}
}

export const cart = new Cart();
