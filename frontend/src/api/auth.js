// Reconstructed application logic from auth-CJlQplIi.js.
import { request } from "@/utils/request.js";
const login = (a, t, d, e) =>
    request("/admin/login", {
      method: "POST",
      data: {
        username: a,
        password: t,
        captchaId: d,
        captchaCode: e,
      },
    }),
  getCaptcha = () => request("/admin/captcha"),
  logout = () =>
    request("/admin/logout", {
      method: "POST",
    }),
  getCurrentAdmin = () => request("/admin/me"),
  updateProfile = (a) =>
    request("/admin/me/profile", {
      method: "PUT",
      data: a,
    }),
  updatePassword = (a) =>
    request("/admin/me/password", {
      method: "PUT",
      data: a,
    }),
  uploadAvatar = (a) => {
    const t = new FormData();
    return (
      t.append("file", a),
      request("/admin/upload/avatar", {
        method: "POST",
        data: t,
      })
    );
  },
  uploadAppLogo = (a) => {
    const t = new FormData();
    return (
      t.append("file", a),
      request("/admin/upload/app-logo", {
        method: "POST",
        data: t,
      })
    );
  };
export {
  getCurrentAdmin,
  logout,
  updateProfile,
  uploadAppLogo,
  uploadAvatar,
  getCaptcha,
  login,
  updatePassword,
};
