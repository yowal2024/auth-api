// Reconstructed application logic from user-DyMy4M2w.js.
import { defineStore } from "pinia";
import { clearToken } from "@/utils/storage.js";
import { setToken } from "@/utils/storage.js";
import { getToken } from "@/utils/storage.js";
import { logout } from "@/api/auth.js";
import { updateProfile } from "@/api/auth.js";
import { getCurrentAdmin } from "@/api/auth.js";
import { login } from "@/api/auth.js";
import { useAppStore } from "@/stores/app.js";
const useUserStore = defineStore("user", {
  state: () => ({
    token: getToken(),
    profile: null,
  }),
  actions: {
    async login(t, e, s, a) {
      const o = await login(t, e, s, a);
      return (
        (this.token = o.token),
        setToken(o.token),
        (this.profile = o.profile),
        o
      );
    },
    async fetchMe() {
      const token = this.token;
      if (!token) return null;
      const profile = await getCurrentAdmin();
      if (this.token === token) this.profile = profile;
      return this.profile;
    },
    async updateProfile(t) {
      const token = this.token;
      const e = await updateProfile(t);
      if (this.token !== token) return e;
      return (
        (this.profile = {
          ...(this.profile || {}),
          ...e,
        }),
        e
      );
    },
    async logout() {
      const token = this.token;
      try {
        await logout();
      } catch {}
      if (this.token === token) this.resetAuth();
    },
    resetAuth() {
      ((this.token = ""), (this.profile = null), clearToken());
      const apps = useAppStore();
      apps.apps = [];
      apps.setCurrent(0);
    },
  },
});
export { useUserStore };
