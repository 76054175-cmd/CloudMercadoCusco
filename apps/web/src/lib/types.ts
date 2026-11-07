import type { RecordModel } from 'pocketbase';
import type { Section, UnitType } from './constants';

export interface Market extends RecordModel {
	name: string;
	slug: string;
	code_prefix: string;
	bubble_icon: string;
	active: boolean;
	hub_location_desc: string;
	open_time: string;
	close_time: string;
	has_own_delivery: boolean;
	delivery_fee: number;
	rappi_enabled: boolean;
}

export interface MetricsSummary {
	quality: number;
	exact_weight: number;
	hygiene: number;
	speed: number;
	overall: number;
	total_votes: number;
}

export interface Stall extends RecordModel {
	market_id: string;
	owner_user_id: string;
	name: string;
	stall_number: string;
	section: Section;
	active: boolean;
	metrics_summary: MetricsSummary | null;
}

export interface CatalogItem extends RecordModel {
	name: string;
	section: Section;
	default_image: string;
}

export interface Product extends RecordModel {
	stall_id: string;
	catalog_item_id: string;
	name: string;
	image: string;
	price: number;
	in_stock: boolean;
	unit_type: UnitType;
	expand?: {
		stall_id?: Stall;
		catalog_item_id?: CatalogItem;
	};
}
