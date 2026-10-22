// Reconstructed application logic from func-Bd29bqIn.js.
import { request } from "@/utils/request.js";
const listFunctions = (n) =>
    request("/admin/function/list", {
      params: n,
    }),
  getFunction = (n) => request(`/admin/function/${n}`),
  saveFunction = (n) =>
    request("/admin/function", {
      method: "POST",
      data: n,
    }),
  deleteFunction = (n) =>
    request(`/admin/function/${n}`, {
      method: "DELETE",
    }),
  runFunction = (n, a, c = {}) =>
    request(`/admin/function/${n || 0}/run`, {
      method: "POST",
      timeout: 35000,
      data: {
        payload: a,
        ...c,
      },
    });
export { getFunction, listFunctions, runFunction, saveFunction, deleteFunction };
