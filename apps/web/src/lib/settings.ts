import { pb } from './pb';
import type { AppSetting } from './types';

let cache: Promise<Map<string, unknown>> | null = null;

/** Lee app_settings una sola vez por sesión (es pública y pequeña). */
function all(): Promise<Map<string, unknown>> {
	cache ??= pb
		.collection('app_settings')
		.getFullList<AppSetting>({ fields: 'key,value' })
		.then((rows) => new Map(rows.map((r) => [r.key, r.value])))
		.catch(() => {
			cache = null;
			return new Map();
		});
	return cache;
}

export async function settingNumber(key: string, fallback: number): Promise<number> {
	const v = (await all()).get(key);
	return typeof v === 'number' ? v : fallback;
}
