<script setup>
import { computed, onMounted, onUnmounted, ref } from "vue";
import { getPosterUrl } from "../config/tmdb";
import SeasonEpisodeList from "./SeasonEpisodeList.vue";

const props = defineProps({
  selected: { type: Object, required: true },
  detail: { type: Object, default: null },
  loading: { type: Boolean, default: false },
  error: { type: String, default: "" },
});
const emit = defineEmits(["close"]);
const closeButton = ref(null);
const backdropUrl = computed(() => getPosterUrl(props.detail?.backdropPath || props.detail?.posterPath || props.selected.posterPath, "w780"));
const posterUrl = computed(() => getPosterUrl(props.detail?.posterPath || props.selected.posterPath, "w342"));
const firstSeasonNumber = computed(() => props.detail?.seasons?.find((season) => season.seasonNumber > 0)?.seasonNumber ?? props.detail?.seasons?.[0]?.seasonNumber);

function labels(value) {
  if (!Array.isArray(value)) return [];
  return value.map((entry) => {
    if (typeof entry === "string") return entry;
    if (!entry || typeof entry !== "object") return "";
    const name = entry.name || entry.title || entry.original_name || "";
    return entry.character ? `${name}（${entry.character}）` : name;
  }).filter(Boolean);
}

function formatRuntime(minutes) {
  if (!minutes) return "";
  const hours = Math.floor(minutes / 60);
  const remainder = minutes % 60;
  return hours ? `${hours} 小时 ${remainder} 分钟` : `${remainder} 分钟`;
}

function closeOnEscape(event) {
  if (event.key === "Escape") emit("close");
}

onMounted(() => {
  window.addEventListener("keydown", closeOnEscape);
  closeButton.value?.focus();
});
onUnmounted(() => window.removeEventListener("keydown", closeOnEscape));
</script>

<template>
  <Teleport to="body">
    <div class="detail-backdrop" @click.self="emit('close')">
      <section class="detail-dialog" role="dialog" aria-modal="true" :aria-label="`${selected.title} 影视详情`">
        <div class="detail-cover" :style="backdropUrl ? { backgroundImage: `linear-gradient(90deg, #111a1df2 0%, #111a1dcc 50%, #111a1d88 100%), url('${backdropUrl}')` } : undefined">
          <button ref="closeButton" class="detail-close" type="button" aria-label="关闭详情" @click="emit('close')">×</button>
          <img v-if="posterUrl" class="detail-poster" :src="posterUrl" :alt="`${selected.title} 海报`" />
          <div class="detail-heading">
            <span class="eyebrow">{{ selected.kind === "movie" ? "MOVIE" : "TV SERIES" }} · TMDB {{ selected.id }}</span>
            <h2>{{ detail?.title || selected.title }}</h2>
            <p v-if="detail?.originalTitle && detail.originalTitle !== detail.title">{{ detail.originalTitle }}</p>
            <p v-if="detail?.tagline" class="detail-tagline">“{{ detail.tagline }}”</p>
            <div class="detail-badges">
              <span>{{ detail?.kind || (selected.kind === "movie" ? "电影" : "电视剧") }}</span>
              <span v-if="detail?.year || selected.releaseDate">{{ detail?.year || selected.releaseDate }}</span>
              <span v-if="detail?.voteAverage">★ {{ detail.voteAverage.toFixed(1) }}</span>
            </div>
          </div>
        </div>
        <div class="detail-body">
          <p v-if="loading" class="detail-message" role="status">正在读取本地片库详情…</p>
          <p v-else-if="error" class="detail-message detail-error" role="alert">{{ error }}</p>
          <template v-else-if="detail">
            <section v-if="detail.overview" class="detail-overview">
              <h3>简介</h3>
              <p>{{ detail.overview }}</p>
            </section>
            <dl class="detail-facts">
              <div v-if="detail.originalLanguage"><dt>原始语言</dt><dd>{{ detail.originalLanguage }}</dd></div>
              <div v-if="labels(detail.originCountries).length"><dt>国家 / 地区</dt><dd>{{ labels(detail.originCountries).join("、") }}</dd></div>
              <div v-if="detail.runtime"><dt>片长</dt><dd>{{ formatRuntime(detail.runtime) }}</dd></div>
              <div v-if="detail.voteCount"><dt>评分人数</dt><dd>{{ detail.voteCount.toLocaleString() }}</dd></div>
              <div v-if="detail.popularity"><dt>热度</dt><dd>{{ detail.popularity.toFixed(1) }}</dd></div>
              <div v-if="detail.status"><dt>状态</dt><dd>{{ detail.status }}{{ detail.inProduction ? " · 正在制作" : "" }}</dd></div>
              <div v-if="detail.kind === '电视剧' && (detail.seasonCount || detail.episodeCount)"><dt>剧集规模</dt><dd>{{ detail.seasonCount }} 季 · {{ detail.episodeCount.toLocaleString() }} 集</dd></div>
              <div v-if="detail.director"><dt>导演</dt><dd>{{ detail.director }}</dd></div>
              <div v-if="labels(detail.networks).length"><dt>电视网</dt><dd>{{ labels(detail.networks).join("、") }}</dd></div>
              <div v-if="labels(detail.companies).length"><dt>制作公司</dt><dd>{{ labels(detail.companies).join("、") }}</dd></div>
              <div v-if="detail.collectionName"><dt>所属系列</dt><dd>{{ detail.collectionName }}</dd></div>
            </dl>
            <section v-if="detail.kind === '电视剧' && detail.seasons?.length" class="detail-season-list">
              <div class="detail-season-heading">
                <h3>分季与分集</h3>
                <span>{{ detail.seasons.length }} 季 · {{ detail.episodeCount.toLocaleString() }} 集</span>
              </div>
              <SeasonEpisodeList
                v-for="season in detail.seasons"
                :key="season.seasonNumber"
                :show-id="detail.id"
                :season="season"
                :initial-open="season.seasonNumber === firstSeasonNumber"
              />
            </section>
            <section v-if="labels(detail.castMembers).length" class="detail-cast">
              <h3>主要演职员</h3>
              <p>{{ labels(detail.castMembers).slice(0, 12).join(" · ") }}</p>
            </section>
            <div class="detail-links">
              <a v-if="detail.imdbId" :href="`https://www.imdb.com/title/${detail.imdbId}/`" target="_blank" rel="noopener noreferrer">IMDb ↗</a>
              <a v-if="detail.homepage" :href="detail.homepage" target="_blank" rel="noopener noreferrer">官方网站 ↗</a>
            </div>
          </template>
        </div>
      </section>
    </div>
  </Teleport>
</template>
