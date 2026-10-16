<script setup>
import { computed, reactive, watch } from "vue";
import { categories } from "../data/categories";
import PosterPreview from "./PosterPreview.vue";
import TmdbSearchPanel from "./TmdbSearchPanel.vue";
import { inferCategory } from "../utils/tmdbMatch";

const props = defineProps({
  item: { type: Object, required: true },
  imageBaseUrl: { type: String, required: true },
  showPoster: { type: Boolean, default: true },
});
const emit = defineEmits(["change"]);
const draft = reactive({
  kind: "电影",
  category: "待确认",
  title: "",
  year: "",
  season: 1,
  episode: 1,
  tmdb_id: 0,
  poster_path: "",
});
const categoryOptions = computed(
  () => categories[draft.kind] || categories["电影"],
);
watch(
  () => props.item,
  (item) => {
    Object.assign(draft, {
      kind: item.kind === "电视剧" ? "电视剧" : "电影",
      category: item.category || "待确认",
      title: item.title || "",
      year: item.year || "",
      season: item.season || 1,
      episode: item.episode || 1,
      tmdb_id: item.tmdb_id || 0,
      poster_path: item.poster_path || "",
    });
  },
  { immediate: true },
);
watch(
  () => draft.kind,
  (kind) => {
    if (!categories[kind].includes(draft.category)) draft.category = "待确认";
    emitChange();
  },
);
function emitChange(seriesCalibration = false) {
  emit("change", {
    source: props.item.source,
    values: { ...draft },
    seriesCalibration,
  });
}
function applyResult(result) {
  draft.title = result.title;
  draft.year = result.year || "";
  draft.tmdb_id = result.id;
  draft.poster_path = result.poster_path || "";
  draft.category = inferCategory(result, draft.kind);
  emitChange(draft.kind === "电视剧");
}
</script>

<template>
  <article class="plan-row" :class="{ 'plan-row-with-poster': showPoster }">
    <PosterPreview
      v-if="showPoster"
      :poster-path="draft.poster_path"
      :image-base-url="imageBaseUrl"
      :alt="`${draft.title} 海报`"
    />
    <div class="plan-paths">
      <span class="source-path">{{ item.source }}</span>
      <span class="arrow">↓ {{ item.destination || item.reason }}</span>
      <div class="plan-editor">
        <select
          v-model="draft.kind"
          aria-label="媒体类型"
          @change="emitChange()"
        >
          <option>电影</option>
          <option>电视剧</option>
        </select>
        <select
          v-model="draft.category"
          aria-label="二级分类"
          @change="emitChange()"
        >
          <option v-for="category in categoryOptions" :key="category">
            {{ category }}
          </option>
        </select>
        <label class="title-edit"
          ><input v-model="draft.title" aria-label="片名" @input="emitChange()"
        /></label>
        <input
          v-model="draft.year"
          aria-label="年份"
          inputmode="numeric"
          placeholder="年份"
          @input="emitChange()"
        />
        <input
          v-if="draft.kind === '电视剧'"
          v-model.number="draft.season"
          aria-label="季"
          type="number"
          min="0"
          max="99"
          @input="emitChange()"
        />
        <input
          v-if="draft.kind === '电视剧'"
          v-model.number="draft.episode"
          aria-label="集"
          type="number"
          min="0"
          max="999"
          @input="emitChange()"
        />
      </div>
      <TmdbSearchPanel
        :source="item.source"
        :kind="draft.kind"
        :title="draft.title"
        :year="draft.year"
        :image-base-url="imageBaseUrl"
        @select="applyResult"
      />
    </div>
    <span class="plan-state" :class="item.status">{{
      {
        ready: "可整理",
        conflict: "重名",
        skipped: "跳过",
        unchanged: "已归档",
      }[item.status] || item.status
    }}</span>
  </article>
</template>
