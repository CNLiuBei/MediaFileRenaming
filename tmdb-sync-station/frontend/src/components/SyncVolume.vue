<script setup>
import { computed } from "vue";
const props = defineProps({ daily: { type: Array, required: true } });
const maxBar = 430;
const total = computed(() => props.daily.reduce((sum, day) => sum + day.count, 0));
</script>

<template>
  <section class="surface volume-card">
    <div class="surface-head"><div><span class="eyebrow">同步吞吐</span><h2>近 7 天更新量</h2></div><span class="volume-total">{{ total.toLocaleString() }} <small>条记录</small></span></div>
    <div class="volume-chart" role="img" aria-label="近七天片库更新量柱状图">
      <div v-for="day in daily" :key="day.day" class="volume-column"><div class="volume-value">{{ day.count }}</div><div class="volume-bar" :style="{ height: `${Math.max(12, day.count / maxBar * 100)}%` }"></div><small>{{ day.day }}</small></div>
      <div v-if="!daily.length" class="empty-state">暂无 PostgreSQL 变更记录</div>
    </div>
    <div class="chart-foot"><span><i></i> 变更条目</span><span>PostgreSQL · 最近 7 天</span></div>
  </section>
</template>
