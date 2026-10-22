// Reconstructed application logic from route-DjOQPZ1X.js.
import { getAdminEntry } from "@/utils/install-status.js";
const entryPath = (t = "") => {
  const r = getAdminEntry();
  return t ? (t.startsWith("/") ? `/${r}${t}` : `/${r}/${t}`) : `/${r}`;
};
export { entryPath };
