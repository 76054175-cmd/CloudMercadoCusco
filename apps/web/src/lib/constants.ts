// Mantener en sync con backend/core/migrations/helpers.go

export const SECTIONS = {
	verduras: { label: 'Verduras', icon: '🥬' },
	carnes_aves: { label: 'Carnes y aves', icon: '🍗' },
	abarrotes: { label: 'Abarrotes', icon: '🛒' },
	quesos_lacteos: { label: 'Quesos y lácteos', icon: '🧀' },
	frutas: { label: 'Frutas', icon: '🍌' },
	otros: { label: 'Otros', icon: '🧺' }
} as const;
export type Section = keyof typeof SECTIONS;

export const UNIT_TYPES = {
	kg: 'kg',
	medio_kg: '½ kg',
	cuarto_kg: '¼ kg',
	atado: 'atado',
	manojo: 'manojo',
	unidad: 'unidad',
	porcion: 'porción'
} as const;
export type UnitType = keyof typeof UNIT_TYPES;

export const formatPEN = (n: number) =>
	new Intl.NumberFormat('es-PE', { style: 'currency', currency: 'PEN' }).format(n);
