<script setup>
defineProps({
  modelValue: { type: String, default: "" },
  status: {
    type: Object,
    default: () => ({ configured: false, available: false, message: "尚未配置同步站" }),
  },
});
const emit = defineEmits(["update:modelValue"]);
</script>

<template>
  <section class="sync-station-panel">
    <label class="tmdb-key-field">
      TMDB 同步站地址（可选）
      <input
        :value="modelValue"
        type="url"
        autocomplete="off"
        placeholder="http://127.0.0.1:8788"
        @input="emit('update:modelValue', $event.target.value)"
      />
    </label>
    <p class="hint">
      同步站与整理应用在同一台 NAS 时填写 http://127.0.0.1:8788。校准会优先搜索同步站的 PostgreSQL 片库，再查本地 SQLite；未命中时才使用 TMDB 在线接口。
    </p>
    <p class="hint" :class="{ 'station-unavailable': status.configured && !status.available }" role="status">
      {{ status.message || "尚未配置同步站" }}
    </p>
  </section>
</template>

<style scoped>
.sync-station-panel {
  margin: 20px 0;
}

.station-unavailable {
  color: #b54708;
}
</style>
