// Reconstructed application logic from useTheme-DMB2o6M0.js.
import { ref } from "vue";
const r = ref(!1);
let e = !1;
const t = () => {
    r.value = document.body.getAttribute("arco-theme") === "dark";
  },
  s = () => {
    if (e) return;
    ((e = !0),
      t(),
      new MutationObserver(t).observe(document.body, {
        attributes: !0,
        attributeFilter: ["arco-theme"],
      }));
  },
  useTheme = () => (
    s(),
    {
      dark: r,
    }
  );
export { useTheme };
