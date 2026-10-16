<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { loadTmdbSeasonDetails } from "../api";
import { getPosterUrl } from "../config/tmdb";

const props = defineProps({
  showId: { type: Number, required: true },
  season: { type: Object, required: true },
  initialOpen: { type: Boolean, default: false },
});

const isExpanded = ref(props.initialOpen);
const isLoading = ref(false);
const hasLoaded = ref(false);
const loadError = ref("");
const tmdbEpisodes = ref(null);
const failedStills = reactive({});

const episodes = computed(() => {
  if (!tmdbEpisodes.value?.length) return props.season.episodes || [];
  const localEpisodes = new Map((props.season.episodes || []).map((item) => [item.episodeNumber, item]));
  return tmdbEpisodes.value.map((episode) => ({
    ...episode,
    title: episode.title || localEpisodes.get(episode.episodeNumber)?.title || "",
  }));
});

async function loadSeason() {
  if (isLoading.value || hasLoaded.value) return;
  isLoading.value = true;
  loadError.value = "";
  try {
    tmdbEpisodes.value = await loadTmdbSeasonDetails({ showId: props.showId, seasonNumber: props.season.seasonNumber });
    hasLoaded.value = true;
  } catch (error) {
    loadError.value = error.message;
  } finally {
    isLoading.value = false;
  }
}

function handleToggle(event) {
  isExpanded.value = event.target.open;
  if (isExpanded.value) loadSeason();
}

function episodeLabel(episodeNumber) {
  const episode = String(episodeNumber).padStart(2, "0");
  return props.season.seasonNumber === 0 ? `SP${episode}` : `S${String(props.season.seasonNumber).padStart(2, "0")}E${episode}`;
}

function stillUrl(stillPath) {
  return getPosterUrl(stillPath, "w300");
}

function formatDate(date) {
  return date ? `播出日期 ${date}` : "播出日期未知";
}

onMounted(() => {
  if (props.initialOpen) loadSeason();
});
</script>

<template>
  <details class="season-episode-group" :open="isExpanded" @toggle="handleToggle">
    <summary>
      <strong>{{ season.name || (season.seasonNumber === 0 ? "特别篇" : `第 ${season.seasonNumber} 季`) }}</strong>
      <span>{{ season.episodeCount }} 集</span>
    </summary>
    <p v-if="isLoading" class="episode-load-note" role="status">正在读取本季集图与详情…</p>
    <div v-if="loadError" class="episode-load-error" role="alert">
      <span>TMDB 集图暂时无法加载，仍显示本地集名。</span>
      <button type="button" @click="loadSeason">重试</button>
    </div>
    <div v-if="episodes.length" class="detail-episode-list">
      <article v-for="episode in episodes" :key="episode.episodeNumber" class="episode-card">
        <img
          v-if="episode.stillPath && !failedStills[episode.episodeNumber]"
          class="episode-still"
          :src="stillUrl(episode.stillPath)"
          :alt="`${episodeLabel(episode.episodeNumber)} ${episode.title || '剧照'}`"
          loading="lazy"
          decoding="async"
          @error="failedStills[episode.episodeNumber] = true"
        />
        <div v-else class="episode-still-fallback" role="img" aria-label="暂无单集剧照">暂无剧照</div>
        <div class="episode-copy">
          <div class="episode-meta">
            <span>{{ episodeLabel(episode.episodeNumber) }}</span>
            <span v-if="episode.voteAverage">★ {{ episode.voteAverage.toFixed(1) }}</span>
          </div>
          <h4>{{ episode.title || `第 ${episode.episodeNumber} 集` }}</h4>
          <p v-if="episode.overview">{{ episode.overview }}</p>
          <small>{{ formatDate(episode.airDate) }}<template v-if="episode.runtime"> · {{ episode.runtime }} 分钟</template></small>
        </div>
      </article>
    </div>
    <p v-else-if="!isLoading" class="detail-message">本地暂未收录这一季的分集标题。</p>
  </details>
</template>
