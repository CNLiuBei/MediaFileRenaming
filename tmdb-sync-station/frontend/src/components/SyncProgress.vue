<script setup>
defineProps({ sync: { type: Object, required: true }, busy: { type: Boolean, default: false } });
const emit = defineEmits(["start"]);
</script>

<template>
  <section class="sync-card">
    <div class="sync-head">
      <div><span class="eyebrow">数据库状态 <i class="live-mark"></i></span><h2>{{ busy ? "正在接收 TMDB 变更" : "PostgreSQL 已连接" }}</h2></div>
      <button class="button sync-button" :disabled="busy || !sync.configured" :title="sync.description" @click="emit('start')"><span>{{ busy ? "同步中…" : sync.configured ? "立即同步" : "同步未配置" }}</span></button>
    </div>
    <div class="film-rail" :class="{ moving: busy, 'is-empty': !sync.lastSync && !busy }" aria-label="同步进度"><span v-for="slot in 44" :key="slot"></span></div>
    <div class="sync-meta">
      <div><small>{{ busy ? "处理进度" : "最近成功同步" }}</small><strong>{{ busy ? `${sync.current.toLocaleString()} / ${sync.total.toLocaleString()} 条` : (sync.lastSync || "暂无同步记录") }}</strong></div>
      <div class="sync-meta-right"><small>{{ busy ? "当前阶段" : "下次自动同步" }}</small><strong>{{ busy ? (sync.stage || sync.description) : sync.nextSync }}</strong></div>
    </div>
    <div v-if="busy" class="progress-track"><span :style="{ width: `${sync.progress}%` }"></span></div>
    <p v-if="sync.state === 'failed'" class="sync-error" role="alert">上次同步失败：{{ sync.description }}</p>
  </section>
</template>
