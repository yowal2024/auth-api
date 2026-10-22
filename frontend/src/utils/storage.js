const TOKEN_KEY = "auth_token";
const CURRENT_APP_KEY = "auth_current_appid";
export const getToken = () => localStorage.getItem(TOKEN_KEY) || "";
export const setToken = (token) => localStorage.setItem(TOKEN_KEY, token);
export const clearToken = () => localStorage.removeItem(TOKEN_KEY);
export const getCurrentAppId = () => Number(localStorage.getItem(CURRENT_APP_KEY) || 0);
export const setCurrentAppId = (id) => localStorage.setItem(CURRENT_APP_KEY, String(id));
