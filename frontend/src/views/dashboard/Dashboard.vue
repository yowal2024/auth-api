<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import { getDashboard } from "@/api/system.js";
import { useTheme } from "@/composables/useTheme.js";
import { reportError } from "@/utils/errors.js";
import {
  watch,
  onMounted,
  onUnmounted,
  markRaw,
  nextTick,
  ref,
  computed,
} from "vue";
import {
  IconUser,
  IconThunderbolt,
  IconBookmark,
  IconCheckCircle,
} from "@arco-design/web-vue/es/icon";
import {
  Row,
  Col,
  Card,
  RadioGroup,
  Radio,
  Timeline,
  TimelineItem,
  Empty,
} from "@arco-design/web-vue";
import { init as iN, use } from "echarts/core";
import { LineChart } from "echarts/charts";
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
use([
  LineChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer,
]);
export default {
  name: "Dashboard",
  components: {
    Row,
    Col,
    Card,
    RadioGroup,
    Radio,
    Timeline,
    TimelineItem,
    Empty,
  },
  setup(r) {
    const { dark: e } = useTheme();
    const t = ref("7");
    const displayedRange = ref("7");
    const a = computed(() => `近 ${displayedRange.value} 天登录与注册趋势`);
    const n = ref([]);
    const i = ref([]);
    const chartRef = ref();
    let s = null;
    let l = null;
    let u = null;
    let requestVersion = 0;
    let disposed = false;
    const handleGetDashboard = async () => {
      const version = ++requestVersion;
      const range = t.value;
      let p;
      try {
        p = await getDashboard({
          range,
        });
      } catch (d) {
        if (!disposed && version === requestVersion) {
          t.value = displayedRange.value;
          reportError(d, "趋势数据加载失败，请重试");
        }
        return;
      }
      if (disposed || version !== requestVersion) return;
      displayedRange.value = range;
      n.value = [
        {
          label: "用户总数",
          value: p.users,
          sub: `今日新增 +${p.usersToday}`,
          color: "blue",
          icon: markRaw(IconUser),
        },
        {
          label: "今日查询量",
          value: p.queryToday ?? 0,
          sub: `API 总调用 ${p.apiTotal ?? 0}`,
          color: "purple",
          icon: markRaw(IconThunderbolt),
        },
        {
          label: "卡密剩余",
          value: p.kamiAvail,
          sub: `今日已用 ${p.kamiUsedToday}`,
          color: "green",
          icon: markRaw(IconBookmark),
        },
        {
          label: "当前在线",
          value: p.onlineNow,
          sub: `今日峰值 ${p.onlinePeakToday}`,
          color: "orange",
          icon: markRaw(IconCheckCircle),
        },
      ];
      i.value = p.recent || [];
      u = p.trend;
      await nextTick();
      if (disposed) return;
      h();
      v();
    };
    const h = () => {
      s ||
        !chartRef.value ||
        ((s = iN(chartRef.value, e.value ? "dark" : null)),
        window.addEventListener("resize", c),
        typeof ResizeObserver < "u" &&
          ((l = new ResizeObserver(() => s?.resize())),
          l.observe(chartRef.value)));
    };
    const v = () => {
      if (!s || !u) return;
      const p = e.value ? "rgba(255,255,255,0.65)" : "#4e5969";
      const d = e.value ? "rgba(255,255,255,0.08)" : "#e5e6eb";
      const g = e.value ? "#232324" : "#ffffff";
      s.setOption(
        {
          backgroundColor: "transparent",
          tooltip: {
            trigger: "axis",
            backgroundColor: g,
            borderColor: d,
            textStyle: {
              color: p,
            },
          },
          grid: {
            left: 40,
            right: 28,
            top: 52,
            bottom: 32,
          },
          legend: {
            data: ["登录", "注册"],
            top: 8,
            right: 8,
            bottom: "auto",
            textStyle: {
              color: p,
            },
          },
          xAxis: {
            type: "category",
            data: u.dates,
            boundaryGap: false,
            axisLine: {
              lineStyle: {
                color: d,
              },
            },
            axisLabel: {
              color: p,
              hideOverlap: true,
            },
            splitLine: {
              lineStyle: {
                color: d,
              },
            },
          },
          yAxis: {
            type: "value",
            minInterval: 1,
            axisLine: {
              lineStyle: {
                color: d,
              },
            },
            axisLabel: {
              color: p,
            },
            splitLine: {
              lineStyle: {
                color: d,
              },
            },
          },
          series: [
            {
              name: "登录",
              type: "line",
              smooth: true,
              data: u.logins,
              areaStyle: {
                opacity: 0.15,
              },
              itemStyle: {
                color: "#165dff",
              },
            },
            {
              name: "注册",
              type: "line",
              smooth: true,
              data: u.regs,
              areaStyle: {
                opacity: 0.15,
              },
              itemStyle: {
                color: "#00b42a",
              },
            },
          ],
        },
        true,
      );
    };
    watch(e, async () => {
      s &&
        (l?.disconnect(),
        (l = null),
        s.dispose(),
        (s = null),
        await nextTick(),
        h(),
        v());
    });
    const c = () => s?.resize();
    onMounted(handleGetDashboard);
    onUnmounted(() => {
      disposed = true;
      window.removeEventListener("resize", c);
      l?.disconnect();
      s?.dispose();
    });
    return {
      e,
      t,
      a,
      n,
      i,
      s,
      l,
      u,
      handleGetDashboard,
      h,
      v,
      c,
      chartRef,
    };
  },
};
</script>
<template>
  <div class="page">
    <Row :gutter="[16, 16]">
      <template #default>
        <template v-for="T in n" :key="T.label">
          <Col :xs="24" :sm="12" :md="12" :lg="6">
            <template #default>
              <div :class="['stat-card', T.color]">
                <div class="icon-wrap">
                  <component :is="T.icon"></component>
                </div>
                <div class="meta">
                  <div class="label">{{ T.label }}</div>
                  <div class="value">{{ T.value }}</div>
                  <div class="sub">
                    <span class="dot"></span>
                    {{ " " + T.sub }}
                  </div>
                </div>
              </div>
            </template>
          </Col>
        </template>
      </template>
    </Row>
    <Row
      :gutter="[16, 16]"
      :style="{
        'margin-top': '16px',
      }"
    >
      <template #default>
        <Col :xs="24" :lg="16">
          <template #default>
            <Card :title="a" :bordered="!1" class="panel-card">
              <template #extra>
                <RadioGroup
                  v-model="t"
                  type="button"
                  size="small"
                  @change="handleGetDashboard"
                >
                  <template #default>
                    <Radio value="7"><template #default>7 天</template></Radio>
                    <Radio value="30"><template #default>30 天</template></Radio>
                  </template>
                </RadioGroup>
              </template>
              <template #default
                ><div ref="chartRef" class="panel-body chart-body"></div
              ></template>
            </Card>
          </template>
        </Col>
        <Col :xs="24" :lg="8">
          <template #default>
            <Card title="最近操作" :bordered="!1" class="panel-card">
              <template #default>
                <div class="panel-body recent-body">
                  <template v-if="i.length">
                    <Timeline :key="0" class="recent-timeline">
                      <template #default>
                        <template v-for="(T, A) in i" :key="A">
                          <TimelineItem :type="T.type">
                            <template #default>
                              <div>{{ T.text }}</div>
                              <div
                                :style="{
                                  color: 'var(--color-text-3)',
                                  'font-size': '12px',
                                  'margin-top': '2px',
                                }"
                              >
                                {{ T.time }}
                              </div>
                            </template>
                          </TimelineItem>
                        </template>
                      </template>
                    </Timeline>
                  </template>
                  <template v-else>
                    <Empty
                      :key="1"
                      description="暂无数据"
                      class="recent-empty"
                    ></Empty>
                  </template>
                </div>
              </template>
            </Card>
          </template>
        </Col>
      </template>
    </Row>
  </div>
</template>
<style scoped>
.panel-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.panel-card :deep(.arco-card-body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.panel-body {
  height: 380px;
}
.chart-body {
  width: 100%;
}
@media (max-width: 540px) {
  .panel-card :deep(.arco-card-header) {
    height: auto;
    min-height: 46px;
    flex-wrap: wrap;
    gap: 8px;
    padding-block: 12px;
  }
  .panel-card :deep(.arco-card-header-title) {
    white-space: normal;
    font-size: 14px;
  }
}
.recent-body {
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.recent-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.recent-timeline :deep(.arco-timeline-item) {
  min-height: auto;
  padding-bottom: 14px;
}
.recent-timeline :deep(.arco-timeline-item:last-child) {
  padding-bottom: 0;
}
.recent-timeline :deep(.arco-timeline-item-content) {
  margin-bottom: 0;
}
</style>
