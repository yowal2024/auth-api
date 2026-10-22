// Reconstructed application logic from format-C45wFN1f.js.

const formatDateTime = (t) => {
    if (!t) return "-";
    const r = typeof t == "number" ? new Date(t * 1e3) : new Date(t),
      n = (e) => String(e).padStart(2, "0");
    return `${r.getFullYear()}-${n(r.getMonth() + 1)}-${n(r.getDate())} ${n(r.getHours())}:${n(r.getMinutes())}:${n(r.getSeconds())}`;
  },
  formatDuration = (t) => {
    if (!t || t <= 0) return "-";
    const r = Math.floor(t / 86400),
      n = Math.floor((t % 86400) / 3600),
      e = Math.floor((t % 3600) / 60);
    return r ? `${r}天${n}小时` : n ? `${n}小时${e}分` : `${e}分钟`;
  },
  formatVIPExpiry = (t) => {
    if (!t) return "未开通";
    if (t >= 9999999999) return "永久";
    const r = Math.floor(Date.now() / 1e3);
    return t <= r ? "已过期" : formatDuration(t - r);
  };
export { formatDateTime, formatVIPExpiry, formatDuration };
