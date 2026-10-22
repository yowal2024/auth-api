import { getInstallStatus } from "@/api/install.js";
const STATUS_KEY = "install_status_v1";
export function getCachedInstallStatus() {
  try {
    const value = sessionStorage.getItem(STATUS_KEY);
    return value ? JSON.parse(value) : null;
  } catch {
    return null;
  }
}
export async function refreshInstallStatus(entry = getCachedInstallStatus()?.adminEntry || "") {
  const response = await getInstallStatus(entry);
  if (response.code === 0 && response.data) {
    sessionStorage.setItem(STATUS_KEY, JSON.stringify(response.data));
    return response.data;
  }
  return null;
}
export const clearInstallStatus = () => sessionStorage.removeItem(STATUS_KEY);
export const getAdminEntry = () => getCachedInstallStatus()?.adminEntry || "admin";
