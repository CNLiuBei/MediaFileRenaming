<script setup>
import MetricCard from "../components/MetricCard.vue";
import RecentChanges from "../components/RecentChanges.vue";
import SyncProgress from "../components/SyncProgress.vue";
import SyncVolume from "../components/SyncVolume.vue";

defineProps({ dashboard: { type: Object, required: true }, busy: Boolean });
defineEmits(["sync", "all-changes"]);
</script>

<template>
  <div class="view-stack">
    <section class="welcome-row">
      <div><span class="eyebrow">你的私人影视目录</span><h1>本地片库，<em>真实入库。</em></h1><p>影视统计、条目检索和同步历史都直接读取 PostgreSQL。</p></div>
      <div class="source-stamp"><span>当前数据源 / 本机</span><strong>PostgreSQL</strong><small>数据库 <i>●</i> 已连接</small></div>
    </section>
    <SyncProgress :sync="dashboard.sync" :busy="busy" @start="$emit('sync')" />
    <section class="metric-grid">
      <MetricCard label="影视条目" :value="dashboard.catalog.titles.toLocaleString()" note="电影 + 电视剧" tone="accent-card" />
      <MetricCard label="电影" :value="dashboard.catalog.movies.toLocaleString()" note="完整影视信息" />
      <MetricCard label="电视剧" :value="dashboard.catalog.series.toLocaleString()" note="含分季与集数" />
      <MetricCard label="分集记录" :value="dashboard.catalog.episodes.toLocaleString()" note="剧集分集索引" />
    </section>
    <section class="lower-grid">
      <RecentChanges :items="dashboard.recent" @all="$emit('all-changes')" />
      <SyncVolume :daily="dashboard.daily" />
    </section>
    <p class="preview-note"><span>◉</span> 已导入 {{ dashboard.catalog.source || '本地片库' }}：{{ dashboard.catalog.titles.toLocaleString() }} 条影视信息、{{ dashboard.catalog.episodes.toLocaleString() }} 集。</p>
  </div>
</template>
