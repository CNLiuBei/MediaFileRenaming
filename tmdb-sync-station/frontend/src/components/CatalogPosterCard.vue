<script setup>
import { computed, ref, watch } from "vue";
import { getPosterUrl } from "../config/tmdb";

const props = defineProps({
  id: { type: Number, required: true },
  title: { type: String, required: true },
  originalTitle: { type: String, required: true },
  kind: { type: String, required: true },
  releaseDate: { type: String, required: true },
  posterPath: { type: String, required: true },
});
const emit = defineEmits(["select"]);

const imageFailed = ref(false);
const imageLoaded = ref(false);
const posterUrl = computed(() => getPosterUrl(props.posterPath));

watch(() => props.posterPath, () => {
  imageFailed.value = false;
  imageLoaded.value = false;
});
</script>

<template>
  <article class="poster-card">
    <button class="poster-trigger" type="button" :aria-label="`查看《${title}》详情`" @click="emit('select', { id, kind: kind === '电影' ? 'movie' : 'tv', title, posterPath, releaseDate })">
      <div class="poster-art">
        <img
          v-if="posterUrl && !imageFailed"
          :src="posterUrl"
          :alt="`${title} 海报`"
          :class="{ 'is-loaded': imageLoaded }"
          loading="lazy"
          decoding="async"
          referrerpolicy="no-referrer"
          @load="imageLoaded = true"
          @error="imageFailed = true"
        />
        <div v-else class="poster-fallback" role="img" :aria-label="`${title} 暂无海报`">
          <span class="fallback-mark">▤</span>
          <strong>{{ title.slice(0, 1) || "影" }}</strong>
          <small>暂无海报</small>
        </div>
        <span class="poster-kind" :class="kind === '电影' ? 'poster-kind-movie' : 'poster-kind-series'">{{ kind }}</span>
        <span v-if="releaseDate" class="poster-year">{{ releaseDate }}</span>
      </div>
    </button>
    <div class="poster-caption">
      <h2 :title="title">{{ title }}</h2>
      <p v-if="originalTitle && originalTitle !== title" :title="originalTitle">{{ originalTitle }}</p>
    </div>
  </article>
</template>
