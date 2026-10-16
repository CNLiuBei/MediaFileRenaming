<script setup>
defineProps({ items: { type: Array, required: true } });
defineEmits(["all"]);
</script>

<template>
  <section class="surface recent-card">
    <div class="surface-head"><div><span class="eyebrow">变更流</span><h2>最近更新</h2></div><button class="text-button" @click="$emit('all')">查看全部 <span>↗</span></button></div>
    <div class="change-list">
      <article v-for="item in items" :key="`${item.kind}-${item.id}`" class="change-row">
        <div class="media-glyph" :class="item.kind === '电影' ? 'movie-glyph' : 'series-glyph'">{{ item.kind === "电影" ? "▰" : "▥" }}</div>
        <div class="change-copy"><strong>{{ item.title }}</strong><small>{{ item.action }} · {{ item.kind }}</small></div>
        <div class="change-time"><small>{{ item.updatedAt }}</small><span>TMDB · {{ item.id }}</span></div>
      </article>
      <div v-if="!items.length" class="empty-state">PostgreSQL 中还没有 TMDB 变更记录。</div>
    </div>
  </section>
</template>
