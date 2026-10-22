// Reconstructed application logic from kami-CPW6rVKt.js.
import { request } from "@/utils/request.js";
const listKamiGroups = () => request("/admin/kami/group/list"),
  saveKamiGroup = (a) =>
    request("/admin/kami/group", {
      method: "POST",
      data: a,
    }),
  deleteKamiGroup = (a) =>
    request(`/admin/kami/group/${a}`, {
      method: "DELETE",
    }),
  listKami = (a) =>
    request("/admin/kami/list", {
      params: a,
    }),
  generateKami = (a) =>
    request("/admin/kami/generate", {
      method: "POST",
      data: a,
    }),
  deleteKami = (a) =>
    request(`/admin/kami/${a}`, {
      method: "DELETE",
    }),
  updateKami = (a, e) =>
    request(`/admin/kami/${a}`, {
      method: "PUT",
      data: e,
    }),
  cleanUsedKami = () =>
    request("/admin/kami/clean-used", {
      method: "POST",
    });
export {
  listKamiGroups,
  saveKamiGroup,
  deleteKami,
  generateKami,
  listKami,
  updateKami,
  deleteKamiGroup,
  cleanUsedKami,
};
