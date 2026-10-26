const TOKEN_KEY = "auth_token";
const CURRENT_APP_KEY = "auth_current_appid";
// Each tab keeps the application displayed by its own app store. Persistence is
// only the initial selection; another tab must not change this tab's requests.
let currentAppId = Number(localStorage.getItem(CURRENT_APP_KEY) || 0);
export const getToken = () => localStorage.getItem(TOKEN_KEY) || "";
export const setToken = (token) => localStorage.setItem(TOKEN_KEY, token);
export const clearToken = () => localStorage.removeItem(TOKEN_KEY);
export const getCurrentAppId = () => currentAppId;
export const setCurrentAppId = (id) => {
  currentAppId = Number(id) || 0;
  localStorage.setItem(CURRENT_APP_KEY, String(currentAppId));
};
