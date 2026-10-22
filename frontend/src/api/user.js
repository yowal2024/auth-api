// Reconstructed application logic from user-wXI6nt9m.js.
import { request } from "@/utils/request.js";
const listUserGroups = () => request("/admin/user/group/list"),
  saveUserGroup = (e) =>
    request("/admin/user/group", {
      method: "POST",
      data: e,
    }),
  deleteUserGroup = (e) =>
    request(`/admin/user/group/${e}`, {
      method: "DELETE",
    }),
  listUsers = (e) =>
    request("/admin/user/list", {
      params: e,
    }),
  getUser = (e) => request(`/admin/user/${e}`),
  createUser = (e) =>
    request("/admin/user", {
      method: "POST",
      data: e,
    }),
  updateUser = (e, t) =>
    request(`/admin/user/${e}`, {
      method: "PUT",
      data: t,
    }),
  deleteUser = (e) =>
    request(`/admin/user/${e}`, {
      method: "DELETE",
    }),
  batchDeleteUsers = (e) =>
    request("/admin/user/batch-delete", {
      method: "POST",
      data: {
        ids: e,
      },
    }),
  exportUsers = async (e) => {
    const response = await request("/admin/user/export", { params: e, responseType: "blob", timeout: 60000 });
    return response.data;
  },
  rewardUsers = (e) =>
    request("/admin/user/reward", {
      method: "POST",
      data: e,
    }),
  listUserExtends = (e) => request(`/admin/user/${e}/extends`),
  deleteUserExtend = (userId, id) => request(`/admin/user/${userId}/extends/${id}`, { method: "DELETE" }),
  saveUserExtends = (e, t) =>
    request(`/admin/user/${e}/extends`, {
      method: "POST",
      data: t,
    }),
  listUserDevices = (e) => request(`/admin/user/${e}/devices`),
  listUserLogs = (e, t) =>
    request(`/admin/user/${e}/logs`, {
      params: t,
    }),
  deleteUserDevice = (e, t) =>
    request(`/admin/user/${e}/devices/${t}`, {
      method: "DELETE",
    }),
  deleteAllUserDevices = (e) =>
    request(`/admin/user/${e}/devices`, {
      method: "DELETE",
    });
export {
  deleteUserExtend,
  createUser,
  deleteUser,
  getUser,
  deleteUserDevice,
  deleteAllUserDevices,
  listUserDevices,
  exportUsers,
  saveUserExtends,
  listUserExtends,
  deleteUserGroup,
  listUserGroups,
  saveUserGroup,
  listUsers,
  listUserLogs,
  rewardUsers,
  updateUser,
  batchDeleteUsers,
};
