<script setup>
import { ref } from "vue";
import { importOfflineCatalog } from "../api";

defineProps({ catalog: { type: Object, default: () => ({ available: false }) } });
const emit = defineEmits(["imported"]);
const selectedFile = ref(null);
const busy = ref(false);
const progress = ref(0);
const error = ref("");
const message = ref("");

function selectFile(event) {
  selectedFile.value = event.target.files?.[0] || null;
  error.value = "";
  message.value = "";
}

async function upload() {
  if (!selectedFile.value || busy.value) return;
  busy.value = true;
  progress.value = 0;
  error.value = "";
  message.value = "正在上传数据库…";
  try {
    const result = await importOfflineCatalog(selectedFile.value, (fraction) => {
      progress.value = Math.round(fraction * 100);
      message.value = `正在上传数据库… ${progress.value}%`;
    });
    emit("imported", result);
    selectedFile.value = null;
    message.value = `导入完成：${result.titles.toLocaleString()} 个影视条目，${result.aliases.toLocaleString()} 个别名。`;
  } catch (cause) {
    error.value = cause.message;
    message.value = "";
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <section class="panel catalog-panel">
    <div class="section-heading">
      <div>
        <p class="eyebrow">离线片库</p>
        <h2>导入完整影视数据库</h2>
      </div>
      <span class="step-dot">03</span>
    </div>
    <p class="muted">
      选择完整的 catalog.sqlite 数据库。文件会直接上传并保存在这台 NAS 的应用私有目录；不会上传到 GitHub。
      自动校准会优先查询离线片库，未命中时再使用已配置的 TMDB API Key。
    </p>
    <p v-if="catalog.available" class="catalog-summary">
      当前片库：{{ Number(catalog.titles || 0).toLocaleString() }} 个影视条目，
      {{ Number(catalog.aliases || 0).toLocaleString() }} 个别名，
      {{ Number(catalog.episodes || 0).toLocaleString() }} 集。
    </p>
    <label class="catalog-file-field">
      SQLite 数据库文件
      <input type="file" accept=".sqlite,.sqlite3,.db,application/vnd.sqlite3" :disabled="busy" @change="selectFile" />
    </label>
    <div v-if="busy" class="catalog-progress" role="progressbar" :aria-valuenow="progress" aria-valuemin="0" aria-valuemax="100">
      <span :style="{ width: `${progress}%` }"></span>
    </div>
    <div class="actions">
      <button class="button primary" :disabled="busy || !selectedFile" @click="upload">
        {{ busy ? "上传中…" : "导入离线片库" }}
      </button>
      <span class="message" role="status">{{ message }}</span>
    </div>
    <p v-if="error" class="message error" role="alert">{{ error }}</p>
    <p class="hint">支持最大 2 GB 的 SQLite 文件；上传期间请保持浏览器页面和网络连接。</p>
  </section>
</template>
