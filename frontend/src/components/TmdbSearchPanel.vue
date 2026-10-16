<script setup>
import { ref, watch } from "vue";
import { searchTmdb } from "../api";
import PosterPreview from "./PosterPreview.vue";

const props = defineProps({
  source: { type: String, required: true },
  kind: { type: String, required: true },
  title: { type: String, default: "" },
  year: { type: String, default: "" },
  imageBaseUrl: { type: String, required: true },
});
const emit = defineEmits(["select"]);
const query = ref(props.title);
const results = ref([]);
const error = ref("");
const searching = ref(false);
watch(
  () => props.title,
  (title) => (query.value = title),
);
async function search() {
  searching.value = true;
  error.value = "";
  results.value = [];
  try {
    results.value = (
      await searchTmdb({
        source: props.source,
        kind: props.kind,
        query: query.value || props.title,
        year: props.year,
      })
    ).results;
  } catch (cause) {
    error.value = cause.message;
  } finally {
    searching.value = false;
  }
}
function select(result) {
  emit("select", result);
  results.value = [];
}
defineExpose({ search });
</script>

<template>
  <div class="tmdb-search-panel">
    <div class="tmdb-tools">
      <input
        v-model="query"
        aria-label="TMDB 搜索词"
        placeholder="输入准确片名搜索 TMDB"
        @keydown.enter="search"
      />
      <button class="button secondary" :disabled="searching" @click="search">
        {{ searching ? "搜索…" : "TMDB 校准" }}
      </button>
    </div>
    <p v-if="error" class="message error" role="alert">{{ error }}</p>
    <div v-if="results.length" class="tmdb-results">
      <button
        v-for="result in results"
        :key="result.id"
        class="tmdb-result"
        @click="select(result)"
      >
        <PosterPreview
          :poster-path="result.poster_path || ''"
          :image-base-url="imageBaseUrl"
          :alt="`${result.title} 海报`"
        />
        <span class="tmdb-result-copy"
          ><strong
            >{{ result.title }}
            <span v-if="result.year">({{ result.year }})</span
            ><span class="tmdb-badge"
              >{{ result.original_language }} ·
              {{ result.origin_country.join(", ") || "地区未知" }}</span
            ></strong
          ><small>{{ result.overview || "暂无简介" }}</small></span
        >
      </button>
    </div>
  </div>
</template>
