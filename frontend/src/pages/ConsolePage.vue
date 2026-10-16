<script setup>
import { computed, ref } from "vue";
import { previewPlan, runPlan } from "../api";
import ApplyActions from "../components/ApplyActions.vue";
import CalibrationActions from "../components/CalibrationActions.vue";
import MediaFormatCards from "../components/MediaFormatCards.vue";
import PlanGroup from "../components/SeriesGroup.vue";
import PlanRow from "../components/PlanRow.vue";
import {
  buildCalibrationOverrides,
  calibrateGroups,
  countCalibrationTargets,
} from "../utils/automaticCalibration";
import { groupPlanItems } from "../utils/planGroups";

const props = defineProps({ config: { type: Object, required: true } });
const items = ref([]);
const overrides = ref({});
const busy = ref(false);
const isAutoCalibrating = ref(false);
const stopAutomaticCalibration = ref(false);
const calibrationProgress = ref("");
const error = ref("");
const message = ref(
  "检查并修正分类、片名、年份或季/集号后再执行；预览不会移动文件。",
);
const readyCount = computed(
  () => items.value.filter((item) => item.status === "ready").length,
);
const groups = computed(() =>
  groupPlanItems(items.value).map((group) => ({
    ...group,
    poster_path:
      group.items
        .map((item) => overrides.value[item.source]?.poster_path)
        .find(Boolean) || "",
  })),
);
const calibrationTargetCount = computed(() =>
  countCalibrationTargets(groups.value),
);
const imageBaseUrl = computed(() => {
  const apiBase = (
    props.config.tmdb_proxy_url || "https://image.tmdb.org"
  ).replace(/\/+$/, "");
  return `${apiBase}/t/p/w185`;
});
function collectOverrides() {
  const result = {};
  for (const item of items.value)
    if (overrides.value[item.source])
      result[item.source] = overrides.value[item.source];
  return result;
}
async function scan() {
  busy.value = true;
  error.value = "";
  let shouldCalibrate = false;
  try {
    const result = await previewPlan(collectOverrides());
    items.value = result.items;
    const ready = result.items.filter((item) => item.status === "ready").length;
    const targetCount = countCalibrationTargets(groupPlanItems(result.items));
    const shouldAutoCalibrate = Boolean(
      props.config.tmdb_key_configured || props.config.offline_catalog?.available,
    );
    shouldCalibrate = shouldAutoCalibrate && targetCount > 0;
    const summary = `扫描完成：${result.count} 个文件，其中 ${ready} 个可整理。`;
    if (shouldCalibrate) {
      message.value = `${summary} 即将自动校准 ${targetCount} 个片名…`;
    } else if (targetCount && !shouldAutoCalibrate) {
      message.value = `${summary} 尚未导入离线片库或配置 TMDB API Key，已跳过自动校准。`;
    } else {
      message.value = summary;
    }
  } catch (cause) {
    error.value = cause.message;
  } finally {
    busy.value = false;
  }
  if (shouldCalibrate) await autoCalibrate();
}
async function autoCalibrate() {
  if (busy.value || !calibrationTargetCount.value) return;
  busy.value = true;
  isAutoCalibrating.value = true;
  stopAutomaticCalibration.value = false;
  error.value = "";
  try {
    const report = await calibrateGroups(
      groups.value,
      (progress) => (calibrationProgress.value = progress),
      () => stopAutomaticCalibration.value,
    );
    const calibrated = buildCalibrationOverrides(report.selections);
    overrides.value = { ...overrides.value, ...calibrated };
    if (Object.keys(calibrated).length) {
      const preview = await previewPlan(collectOverrides());
      items.value = preview.items;
    }
    const stopped = report.cancelled ? "已停止；" : "";
    message.value = `${stopped}自动校准完成：匹配 ${report.selections.length} 组，未找到高置信结果 ${report.unmatched} 组。请核对预览后再执行整理。`;
    if (report.error) error.value = `校准中断：${report.error}`;
  } catch (cause) {
    error.value = cause.message;
  } finally {
    busy.value = false;
    isAutoCalibrating.value = false;
    calibrationProgress.value = "";
  }
}
function requestStopCalibration() {
  stopAutomaticCalibration.value = true;
  calibrationProgress.value = "当前 TMDB 请求完成后停止…";
}
async function execute() {
  if (
    !window.confirm(
      `确认整理 ${readyCount.value} 个文件？文件会移动到预览路径。`,
    )
  )
    return;
  busy.value = true;
  error.value = "";
  try {
    const result = await runPlan(collectOverrides());
    overrides.value = {};
    message.value = `整理完成：移动 ${result.moved} 个，跳过 ${result.skipped} 个，失败 ${result.failed} 个。`;
    await scan();
  } catch (cause) {
    error.value = cause.message;
  } finally {
    busy.value = false;
  }
}
function saveOverride(payload) {
  const group = groups.value.find(
    (candidate) =>
      candidate.isSeries &&
      candidate.items.some((item) => item.source === payload.source),
  );
  if (!payload.seriesCalibration || !group) {
    overrides.value[payload.source] = payload.values;
    return;
  }
  const seriesValues = { ...payload.values };
  delete seriesValues.season;
  delete seriesValues.episode;
  for (const item of group.items) {
    overrides.value[item.source] = {
      ...seriesValues,
      kind: "电视剧",
      season: item.season || 1,
      episode: item.episode || 1,
    };
  }
}
</script>
<template>
  <section class="intro">
    <div>
      <p class="eyebrow">媒体库工作台</p>
      <h2>让电影和剧集各归其位</h2>
      <p class="muted">扫描后优先用离线片库校准影视信息；核对预览与分类后，再执行整理。</p>
    </div>
    <div class="intro-icon" aria-hidden="true">▤</div>
  </section>
  <section class="panel preview-panel">
    <div class="section-heading">
      <div>
        <p class="eyebrow">媒体整理</p>
        <h2>预览整理方案</h2>
      </div>
      <span class="step-dot">01</span>
    </div>
    <MediaFormatCards />
    <CalibrationActions
      :busy="busy"
      :has-items="items.length > 0"
      :target-count="calibrationTargetCount"
      :is-calibrating="isAutoCalibrating"
      :progress="calibrationProgress"
      :message="message"
      @scan="scan"
      @calibrate="autoCalibrate"
      @stop="requestStopCalibration"
    />
    <p v-if="items.length" class="hint-inline">
      扫描完成后优先查询离线片库；在线 TMDB 请求按 0.8 QPS 限速。只采用高置信匹配；校准仅更新预览，不会移动文件。
    </p>
    <p v-if="error" class="message error" role="alert">{{ error }}</p>
    <div v-if="items.length" class="plan-wrap">
      <template v-for="group in groups" :key="group.key">
        <PlanGroup
          v-if="group.isSeries"
          :group="group"
          :image-base-url="imageBaseUrl"
          @change="saveOverride"
        />
        <PlanRow
          v-else
          :item="group.items[0]"
          :image-base-url="imageBaseUrl"
          @change="saveOverride"
        />
      </template>
    </div>
    <ApplyActions :ready-count="readyCount" :busy="busy" @execute="execute" />
  </section>
</template>
