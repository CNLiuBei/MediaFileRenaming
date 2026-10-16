<script setup>
defineProps({ dashboard: { type: Object, required: true } });
const labels = { success: "成功", failed: "失败", running: "进行中" };
</script>

<template>
  <div class="view-stack">
    <section class="page-heading"><div><span class="eyebrow">SYNC / HISTORY</span><h1>同步记录</h1><p>查看最近的同步任务与更新条目。</p></div><span class="data-pill">最近 30 天</span></section>
    <section class="surface history-surface"><div v-if="!dashboard.runs.length" class="empty-state">PostgreSQL 中还没有同步任务记录。</div><div v-for="run in dashboard.runs" :key="run.id" class="run-row"><div class="run-marker">{{ run.state === 'success' ? '✓' : run.state === 'failed' ? '!' : '↻' }}</div><div class="run-copy"><strong>{{ labels[run.state] || run.state }} · 同步任务 #{{ run.id }}</strong><small>{{ run.stage || '等待处理' }} · {{ run.current.toLocaleString() }} / {{ run.total.toLocaleString() }} 条{{ run.error ? ` · ${run.error}` : '' }}</small></div><span class="run-state">{{ labels[run.state] || run.state }}</span><time>{{ run.startedAt }}</time></div></section>
    <section class="surface recent-card"><div class="surface-head"><div><span class="eyebrow">最近入库</span><h2>变更条目</h2></div></div><div v-if="!dashboard.recent.length" class="empty-state">数据库中还没有 TMDB 变更记录。</div><div class="change-list"><article v-for="item in dashboard.recent" :key="`${item.kind}-${item.id}`" class="change-row"><div class="media-glyph" :class="item.kind === '电影' ? 'movie-glyph' : 'series-glyph'">{{ item.kind === '电影' ? '▰' : '▥' }}</div><div class="change-copy"><strong>{{ item.title }}</strong><small>{{ item.action }} · {{ item.kind }}</small></div><div class="change-time"><small>{{ item.updatedAt }}</small><span>TMDB · {{ item.id }}</span></div></article></div></section>
    <p class="preview-note"><span>◉</span> 同步任务和变更记录均从 PostgreSQL 读取。</p>
  </div>
</template>
