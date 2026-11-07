import PocketBase from 'pocketbase';

// Mismo origen siempre: en dev Vite hace proxy de /api al core; en prod la web está embebida.
export const pb = new PocketBase(window.location.origin);
pb.autoCancellation(false);
