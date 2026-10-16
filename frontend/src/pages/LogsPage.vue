<script setup>
import { computed, onActivated, ref } from "vue";
import { getLogs } from "../api";

const items = ref([]);
const error = ref("");
const loading = ref(false);
const selectedAction = ref("全部");
const actions = ["全部", "扫描", "手动整理", "自动整理", "错误"];
const visibleItems = computed(() =>
  selectedAction.value === "全部"
    ? items.value
    : items.value.filter((entry) => entry.action === selectedAction.value),
);
const actionCounts = computed(() =>
  Object.fromEntries([
    ["全部", items.value.length],
    ...actions.slice(1).map((action) => [
      action,
      items.value.filter((entry) => entry.action === action).length,
    ]),
  ]),
);
const totals = computed(() =>
  items.value.reduce(
    (result, entry) => ({
      scanned: result.scanned + (Number(entry.total) || 0),
      ready: result.ready + (Number(entry.ready) || 0),
      moved: result.moved + (Number(entry.moved) || 0),
      failed: result.failed + (Number(entry.failed) || 0),
    }),
    { scanned: 0, ready: 0, moved: 0, failed: 0 },
  ),
);

async function refresh() {
  loading.value = true;
  error.value = "";
  try {
    items.value = (await getLogs()).items;
  } catch (cause) {
    error.value = cause.message;
  } finally {
    loading.value = false;
  }
}

function statusLabel(status) {
  const labels = {
    ready: "待整理",
    skipped: "已跳过",
    conflict: "目标冲突",
    unchanged: "未变化",
    failed: "失败",
  };
  return labels[status] || status;
}

onActivated(refresh);
</script>

<template>
  <section class="panel logs-panel">
    <div class="section-heading">
      <div>
        <p class="eyebrow">操作记录</p>
        <h2>整理日志</h2>
      </div>
      <button class="button secondary" :disabled="loading" @click="refresh">
        {{ loading ? "读取中…" : "刷新" }}
      </button>
    </div>
    <p class="muted">扫描、手动整理、自动整理和异常记录，最多保留最近 500 条。</p>
    <div class="log-filters" role="group" aria-label="按操作类型筛选">
      <button
        v-for="action in actions"
        :key="action"
        class="log-filter"
        :class="{ active: selectedAction === action }"
        type="button"
        :aria-pressed="selectedAction === action"
        @click="selectedAction = action"
      >
        {{ action }}<span class="log-filter-count">{{ actionCounts[action] }}</span>
      </button>
    </div>
    <div class="log-summary">
      <span>已读取 {{ items.length }} 条记录</span>
      <span>扫描 {{ totals.scanned }} 个 · 可整理 {{ totals.ready }} 个 · 移动 {{ totals.moved }} 个 · 失败 {{ totals.failed }} 个</span>
    </div>
    <p v-if="error" class="message error" role="alert">{{ error }}</p>
    <div class="logs-list" aria-live="polite">
      <div v-if="!visibleItems.length && !loading && !error" class="logs-empty">
        {{ items.length ? "没有符合条件的日志" : "暂无活动记录" }}
      </div>
      <article
        v-for="(entry, index) in visibleItems"
        :key="`${entry.time}-${index}`"
        class="log-entry"
      >
        <div
          class="log-symbol"
          :class="{ failed: entry.action === '错误' || entry.failed > 0 }"
          aria-hidden="true"
        >
          {{ entry.action === "错误" || entry.failed > 0 ? "!" : "✓" }}
        </div>
        <div class="log-content">
          <strong>{{ entry.action }}</strong>
          <p>{{ entry.message }}</p>
          <div v-if="entry.total || entry.moved || entry.skipped || entry.failed" class="log-counts">
            <span>扫描 {{ entry.total || 0 }}</span>
            <span>移动 {{ entry.moved || 0 }}</span>
            <span>跳过 {{ entry.skipped || 0 }}</span>
            <span :class="{ 'failed-count': entry.failed > 0 }">失败 {{ entry.failed || 0 }}</span>
          </div>
          <div v-if="entry.source_dirs?.length || entry.destination_dir" class="log-paths">
            <span v-for="source in entry.source_dirs || []" :key="source">来源：{{ source }}</span>
            <span v-if="entry.destination_dir">目标：{{ entry.destination_dir }}</span>
          </div>
          <details v-if="entry.details?.length" class="log-details">
            <summary>查看未整理详情（{{ entry.details_total ?? entry.details.length }} 项）</summary>
            <ul>
              <li v-for="(detail, detailIndex) in entry.details" :key="`${detail.source}-${detailIndex}`">
                <strong>{{ statusLabel(detail.status) }}：{{ detail.message }}</strong>
                <code>{{ detail.source }}</code>
                <code v-if="detail.destination">目标：{{ detail.destination }}</code>
              </li>
            </ul>
            <small v-if="entry.details_total > entry.details.length">
              为控制日志大小，仅显示前 {{ entry.details.length }} 项。
            </small>
          </details>
          <small>{{ entry.time }}</small>
        </div>
      </article>
    </div>
  </section>
</template>
