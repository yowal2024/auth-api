// Reconstructed application logic from app-Dhmbjjup.js.

import { request } from "@/utils/request.js";
const listApps = () => request("/admin/app/list"),
  createApp = (e) =>
    request("/admin/app", {
      method: "POST",
      data: e,
    }),
  updateApp = (e, n) =>
    request(`/admin/app/${e}`, {
      method: "PUT",
      data: n,
    }),
  deleteApp = (e) =>
    request(`/admin/app/${e}`, {
      method: "DELETE",
    }),
  getMTLS = (e) => request(`/admin/app/${e}/mtls`),
  generateCA = (e) =>
    request(`/admin/app/${e}/mtls/gen-ca`, {
      method: "POST",
    }),
  issueClientCertificate = (e) =>
    request(`/admin/app/${e}/mtls/issue-cert`, {
      method: "POST",
    }),
  listNotices = (e) =>
    request("/admin/notice/list", {
      params: e,
    }),
  saveNotice = (e) =>
    request("/admin/notice", {
      method: "POST",
      data: e,
    }),
  deleteNotice = (e) =>
    request(`/admin/notice/${e}`, {
      method: "DELETE",
    }),
  listExtends = (e) =>
    request("/admin/extend/list", {
      params: e,
    }),
  saveExtend = (e) =>
    request("/admin/extend", {
      method: "POST",
      data: e,
    }),
  deleteExtend = (e) =>
    request(`/admin/extend/${e}`, {
      method: "DELETE",
    }),
  listExtendFiles = (e) =>
    request("/admin/extend-file/list", {
      params: e,
    }),
  listFileCategories = (e) =>
    request("/admin/extend-file/categories", {
      params: e,
    }),
  uploadExtendFile = (e, n, s) =>
    request("/admin/extend-file/upload", {
      method: "POST",
      params: e,
      data: n,
      timeout: 3e5,
      onUploadProgress: s,
    }),
  updateExtendFile = (e, n) =>
    request(`/admin/extend-file/${e}`, {
      method: "PUT",
      data: n,
    }),
  deleteExtendFile = (e) =>
    request(`/admin/extend-file/${e}`, {
      method: "DELETE",
    }),
  listVersions = (e) =>
    request("/admin/ver/list", {
      params: e,
    }),
  saveVersion = (e) =>
    request("/admin/ver", {
      method: "POST",
      data: e,
    }),
  deleteVersion = (e) =>
    request(`/admin/ver/${e}`, {
      method: "DELETE",
    }),
  uploadAPK = (e) => {
    const n = new FormData();
    return (
      n.append("file", e),
      request("/admin/ver/upload-apk", {
        method: "POST",
        data: n,
        timeout: 300000,
      })
    );
  },
  listBlocks = (e) =>
    request("/admin/block/list", {
      params: e,
    }),
  createBlock = (e) =>
    request("/admin/block", {
      method: "POST",
      data: e,
    }),
  deleteBlock = (e) =>
    request(`/admin/block/${e}`, {
      method: "DELETE",
    }),
  listEncryptions = (e) =>
    request("/admin/encrypt/list", {
      params: e,
    }),
  saveEncryption = (e) =>
    request("/admin/encrypt", {
      method: "POST",
      data: e,
    }),
  deleteEncryption = (e) =>
    request(`/admin/encrypt/${e}`, {
      method: "DELETE",
    }),
  generateRSAKey = () => request("/admin/encrypt/gen-rsa-key");
export {
  deleteVersion,
  listVersions,
  saveVersion,
  uploadAPK,
  createApp,
  deleteApp,
  listApps,
  updateApp,
  createBlock,
  deleteBlock,
  listBlocks,
  deleteEncryption,
  generateRSAKey,
  listEncryptions,
  saveEncryption,
  deleteExtend,
  listFileCategories,
  deleteExtendFile,
  listExtendFiles,
  updateExtendFile,
  uploadExtendFile,
  listExtends,
  saveExtend,
  generateCA,
  getMTLS,
  issueClientCertificate,
  deleteNotice,
  listNotices,
  saveNotice,
};
