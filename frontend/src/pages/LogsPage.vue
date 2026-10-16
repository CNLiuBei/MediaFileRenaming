<script setup>
import { onMounted, ref } from "vue";
import { getLogs } from "../api";
const items = ref([]);
const error = ref("");
const loading = ref(false);
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
onMounted(refresh);
</script>
<template>
  <section class="panel logs-panel">
    <div class="section-heading">
      <div>
        <p class="eyebrow">操作记录</p>
        <h2>最近活动</h2>
      </div>
      <button class="button secondary" :disabled="loading" @click="refresh">
        {{ loading ? "读取中…" : "刷新" }}
      </button>
    </div>
    <p class="muted">记录扫描、手动整理和自动整理的结果。</p>
    <p v-if="error" class="message error" role="alert">{{ error }}</p>
    <div class="logs-list" aria-live="polite">
      <div v-if="!items.length && !loading && !error" class="logs-empty">
        暂无活动记录
      </div>
      <article
        v-for="(entry, index) in items"
        :key="`${entry.time}-${index}`"
        class="log-entry"
      >
        <div
          class="log-symbol"
          :class="entry.action === '错误' ? 'failed' : ''"
        >
          {{ entry.action === "错误" ? "!" : "✓" }}
        </div>
        <div>
          <strong>{{ entry.action }}</strong>
          <p>{{ entry.message }}</p>
          <small>{{ entry.time }}</small>
        </div>
      </article>
    </div>
  </section>
</template>
