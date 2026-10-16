<script setup>
import { computed } from "vue";
const props = defineProps({ database: { type: Object, default: null }, dashboard: { type: Object, required: true } });
const sync = computed(() => props.dashboard.sync);
</script>

<template>
  <div class="view-stack">
    <section class="page-heading"><div><span class="eyebrow">PREFERENCES / CONNECTION</span><h1>连接与设置</h1><p>查看数据库连接和 TMDB 同步配置状态。</p></div><span class="data-pill"><i></i> 本机服务</span></section>
    <section class="surface settings-surface">
      <div class="settings-section">
        <div class="settings-copy"><span class="eyebrow">数据库</span><h2>PostgreSQL</h2><p>同步站启动时会连接数据库并自动应用结构迁移。</p></div>
        <div class="settings-fields">
          <label>连接状态<input :value="database ? '已连接' : '未连接'" readonly /><small>连接地址由服务端 DATABASE_URL 环境变量提供</small></label>
          <label>数据库名称<input :value="database?.database || '等待连接'" readonly /><small>{{ database?.user ? `当前数据库用户：${database.user}` : '检查本地 PostgreSQL 服务状态' }}</small></label>
        </div>
      </div>
      <div class="settings-section">
        <div class="settings-copy"><span class="eyebrow">数据源</span><h2>TMDB API</h2><p>服务启动后自动运行增量同步，也可以在总览页手动启动。</p></div>
        <div class="settings-fields">
          <label>API 凭据<input :value="sync.authConfigured ? '已通过服务端环境变量配置' : '使用代理侧凭据 / 未设置'" readonly /><small>可通过 TMDB_API_KEY 或 TMDB_READ_ACCESS_TOKEN 设置，不会发送到浏览器</small></label>
          <label>代理地址<input :value="sync.apiBase || '未配置'" readonly /><small>TMDB_SYNC_QPS：{{ sync.rateLimit || 20 }} 次 / 秒</small></label>
        </div>
      </div>
      <div class="settings-footer"><span class="save-confirm">{{ sync.configured ? '同步服务已配置' : '请配置 TMDB_API_BASE_URL' }}</span><span>{{ sync.autoSync ? '每日自动同步' : '自动同步已关闭' }} · 失败后水位不前移</span></div>
    </section>
    <section class="notice-card"><span>i</span><p><strong>离线片库快照</strong><br />来源：{{ dashboard.catalog.source || '未记录' }} · 影视 {{ dashboard.catalog.titles.toLocaleString() }} 条 · 别名 {{ dashboard.catalog.aliases.toLocaleString() }} 条 · 分集 {{ dashboard.catalog.episodes.toLocaleString() }} 集<br />源库更新时间：{{ dashboard.catalog.sourceUpdatedAt || '未知' }} · 本机导入时间：{{ dashboard.catalog.importedAt || '未知' }}<br />数据库：{{ database?.database || '连接不可用' }} · {{ database?.version || 'PostgreSQL 状态不可用' }}</p></section>
  </div>
</template>
