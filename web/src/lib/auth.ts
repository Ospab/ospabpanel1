export const TOKEN_KEY = 'ospab_token';
export const USER_KEY = 'ospab_user';

export function saveAuth(token: string, user: any){ localStorage.setItem(TOKEN_KEY, token); localStorage.setItem(USER_KEY, JSON.stringify(user)); }
export function getToken(){ return localStorage.getItem(TOKEN_KEY); }
export function getUser(){ const raw = localStorage.getItem(USER_KEY); return raw? JSON.parse(raw): null; }
export function clearToken(){ localStorage.removeItem(TOKEN_KEY); localStorage.removeItem(USER_KEY); }
