<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { onMounted, onUnmounted, computed, ref } from "vue";
import { IconRight, IconCheck } from "@arco-design/web-vue/es/icon";
const F = 40;
export default {
  props: {
    passed: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["pass", "reset"],
  name: "SliderBar",
  components: {
    IconCheck,
    IconRight,
  },
  setup(y, { emit: B }) {
    const props = y;
    const emit = B;
    const trackRef = ref(null);
    const dragging = ref(false);
    const dragOffset = ref(0);
    const dragStartX = ref(0);
    const trackWidth = ref(0);
    const maxOffset = computed(() => Math.max(0, trackWidth.value - F));
    const thumbLeft = computed(() =>
      props.passed ? maxOffset.value + "px" : Math.min(dragOffset.value, maxOffset.value) + "px",
    );
    const fillWidth = computed(() =>
      props.passed ? "100%" : Math.min(dragOffset.value, maxOffset.value) + F + "px",
    );
    const measureTrack = () => {
      trackRef.value && (trackWidth.value = trackRef.value.clientWidth);
    };
    const startDrag = (s) => {
      props.passed ||
        (measureTrack(),
        (dragging.value = true),
        (dragStartX.value = s.type === "touchstart" ? s.touches[0].clientX : s.clientX),
        (dragOffset.value = 0),
        document.addEventListener("mousemove", moveDrag),
        document.addEventListener("mouseup", stopDrag),
        document.addEventListener("touchmove", moveDrag, {
          passive: false,
        }),
        document.addEventListener("touchend", stopDrag));
    };
    const moveDrag = (s) => {
      if (!dragging.value) return;
      s.cancelable && s.preventDefault();
      const n = s.type === "touchmove" ? s.touches[0].clientX : s.clientX;
      dragOffset.value = Math.max(0, n - dragStartX.value);
    };
    const stopDrag = () => {
      dragging.value &&
        ((dragging.value = false),
        dragOffset.value >= maxOffset.value - 5
          ? ((dragOffset.value = maxOffset.value), emit("pass"))
          : ((dragOffset.value = 0), emit("reset")),
        document.removeEventListener("mousemove", moveDrag),
        document.removeEventListener("mouseup", stopDrag),
        document.removeEventListener("touchmove", moveDrag),
        document.removeEventListener("touchend", stopDrag));
    };
    onMounted(() => {
      measureTrack();
      window.addEventListener("resize", measureTrack);
    });
    onUnmounted(() => {
      window.removeEventListener("resize", measureTrack);
      document.removeEventListener("mousemove", moveDrag);
      document.removeEventListener("mouseup", stopDrag);
      document.removeEventListener("touchmove", moveDrag);
      document.removeEventListener("touchend", stopDrag);
    });
    return {
      B,
      props,
      emit,
      dragging,
      dragOffset,
      dragStartX,
      trackWidth,
      maxOffset,
      thumbLeft,
      fillWidth,
      measureTrack,
      startDrag,
      moveDrag,
      stopDrag,
      trackRef,
    };
  },
};
</script>
<template>
  <div
    :class="[
      'slider-bar',
      {
        passed: passed,
        dragging: dragging,
      },
    ]"
  >
    <div ref="trackRef" class="slider-track">
      <div
        class="slider-fill"
        :style="{
          width: fillWidth,
        }"
      ></div>
      <span class="slider-text">
        {{ passed ? "验证通过" : dragging ? "" : "向右滑动完成验证" }}
      </span>
      <div
        class="slider-thumb"
        :style="{
          left: thumbLeft,
        }"
        @mousedown="startDrag"
        @touchstart.passive="startDrag"
      >
        <template v-if="passed"><IconCheck :key="1"></IconCheck></template>
        <template v-else><IconRight :key="0"></IconRight></template>
      </div>
    </div>
  </div>
</template>
<style scoped>
.slider-bar {
  width: 100%;
  user-select: none;
}
.slider-track {
  position: relative;
  height: 40px;
  border-radius: 20px;
  background: #f0f1f5;
  border: 1px solid var(--color-border-2);
  overflow: hidden;
}
.slider-fill {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  background: #e8f3ff;
  border-radius: 20px;
  transition: width 0.05s linear;
  pointer-events: none;
}
.passed .slider-fill {
  background: #e8ffea;
  transition: width 0.2s ease;
}
.slider-text {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: var(--color-text-3);
  pointer-events: none;
  z-index: 1;
}
.passed .slider-text {
  color: #00b42a;
  font-weight: 500;
}
.slider-thumb {
  position: absolute;
  top: 0;
  left: 0;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #fff;
  border: 1px solid var(--color-border-2);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: grab;
  z-index: 2;
  box-shadow: 0 2px 6px #0000001f;
  transition:
    left 0s,
    background 0.2s;
  color: #165dff;
  font-size: 16px;
}
.slider-thumb:active {
  cursor: grabbing;
}
.passed .slider-thumb {
  background: #00b42a;
  border-color: #00b42a;
  color: #fff;
  cursor: default;
}
.dragging .slider-thumb {
  transition: none;
}
</style>
