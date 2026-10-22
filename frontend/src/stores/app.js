import { defineStore } from "pinia";
import { getCurrentAppId, setCurrentAppId, getToken } from "@/utils/storage.js";
import { listApps } from "@/api/app.js";
export const useAppStore = defineStore("app", {
  state: () => ({
    apps: [],
    currentId: getCurrentAppId(),
    fetchSequence: 0,
  }),
  getters: {
    current(e) {
      return e.apps.find((n) => n.id === e.currentId) || null;
    },
  },
  actions: {
    async fetchApps() {
      const sequence = ++this.fetchSequence;
      const token = getToken();
      const apps = await listApps();
      if (sequence !== this.fetchSequence || token !== getToken())
        return this.apps;
      this.apps = apps;
      if (!this.apps.some((app) => app.id === this.currentId))
        this.setCurrent(this.apps[0]?.id || 0);
      return this.apps;
    },
    setCurrent(e) {
      this.currentId = Number(e) || 0;
      setCurrentAppId(this.currentId);
    },
  },
});
