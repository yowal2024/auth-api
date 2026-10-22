// Reconstructed application logic from system-Cyqo6nkk.js.
import { request } from "@/utils/request.js";
const listAdmins = () => request("/admin/admin/list"),
  saveAdmin = (s) =>
    request("/admin/admin", {
      method: "POST",
      data: s,
    }),
  deleteAdmin = (s) =>
    request(`/admin/admin/${s}`, {
      method: "DELETE",
    }),
  getSystemSettings = () => request("/admin/system/settings"),
  saveSystemSettings = (s) =>
    request("/admin/system/settings", {
      method: "POST",
      data: s,
    }),
  getHTTPSSettings = () => request("/admin/system/https"),
  saveHTTPSSettings = (s) =>
    request("/admin/system/https", {
      method: "POST",
      data: s,
    }),
  generateHTTPSCertificate = (s) =>
    request("/admin/system/https/gen-cert", {
      method: "POST",
      data: s,
    }),
  listAPIRoutes = (s) =>
    request("/admin/system/api-routes", {
      params: s,
    }),
  saveAPIRoutes = (s) =>
    request("/admin/system/api-routes/save", {
      method: "POST",
      data: {
        items: s,
      },
    }),
  randomizeAPIRoutes = () =>
    request("/admin/system/api-routes/randomize-all", {
      method: "POST",
    }),
  listAPICodes = (s) =>
    request("/admin/system/api-codes", {
      params: s,
    }),
  saveAPICodes = (s) =>
    request("/admin/system/api-codes/save", {
      method: "POST",
      data: {
        items: s,
      },
    }),
  randomizeAPICodes = () =>
    request("/admin/system/api-codes/randomize-all", {
      method: "POST",
    }),
  listLogs = (s) =>
    request("/admin/log/list", {
      params: s,
    }),
  listLogActions = (s) =>
    request("/admin/log/actions", {
      params: s,
    }),
  clearLogs = (s) =>
    request("/admin/log/clear", {
      method: "POST",
      data: s,
    }),
  getDashboard = (s) =>
    request("/admin/dashboard", {
      params: s,
    });
export {
  deleteAdmin,
  listAdmins,
  saveAdmin,
  saveAPICodes,
  listAPICodes,
  randomizeAPICodes,
  saveAPIRoutes,
  listAPIRoutes,
  randomizeAPIRoutes,
  getDashboard,
  generateHTTPSCertificate,
  getHTTPSSettings,
  saveHTTPSSettings,
  listLogActions,
  clearLogs,
  listLogs,
  saveSystemSettings,
  getSystemSettings,
};
