<script setup>
defineProps({ label: { type: String, required: true }, value: { type: String, required: true }, note: { type: String, default: "" }, tone: { type: String, default: "" } });
</script>

<template>
  <article class="metric-card" :class="tone">
    <div class="metric-top"><span>{{ label }}</span><span class="metric-spark">⌁</span></div>
    <strong>{{ value }}</strong>
    <small>{{ note }}</small>
  </article>
</template>
