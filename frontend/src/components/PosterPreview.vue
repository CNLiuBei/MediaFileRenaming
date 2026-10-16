<script setup>
import { computed, ref, watch } from "vue";

const props = defineProps({
  posterPath: { type: String, default: "" },
  imageBaseUrl: { type: String, required: true },
  alt: { type: String, default: "影视海报" },
  size: { type: String, default: "small" },
});
const failed = ref(false);
const imageUrl = computed(() => {
  if (!props.posterPath) return "";
  return `${props.imageBaseUrl.replace(/\/+$/, "")}/${props.posterPath.replace(/^\/+/, "")}`;
});
watch(imageUrl, () => (failed.value = false));
</script>

<template>
  <img
    v-if="imageUrl && !failed"
    class="poster-preview"
    :class="`poster-${size}`"
    :src="imageUrl"
    :alt="alt"
    loading="lazy"
    @error="failed = true"
  />
  <div
    v-else
    class="poster-placeholder"
    :class="`poster-${size}`"
    aria-label="暂无海报"
  >
    <span>暂无<br />海报</span>
  </div>
</template>
