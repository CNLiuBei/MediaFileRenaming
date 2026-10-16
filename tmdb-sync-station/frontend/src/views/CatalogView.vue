<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { loadCatalog, loadCatalogDetail } from "../api";
import CatalogPosterCard from "../components/CatalogPosterCard.vue";
import CatalogDetailDialog from "../components/CatalogDetailDialog.vue";

const props = defineProps({ dashboard: { type: Object, required: true } });
const filters = ["全部", "电影", "电视剧"];
const posterSkeletons = Array.from({ length: 12 }, (_, index) => index);
const query = ref("");
const filter = ref("全部");
const rows = ref([]);
const hasMore = ref(false);
const loading = ref(false);
const loadError = ref("");
const selectedTitle = ref(null);
const selectedDetail = ref(null);
const detailLoading = ref(false);
const detailError = ref("");
let latestRequestId = 0;
let detailRequestId = 0;
async function refreshCatalog(append = false) {
  const requestId = ++latestRequestId;
  const offset = append ? rows.value.length : 0;
  if (!append) {
    rows.value = [];
    hasMore.value = false;
  }
  loading.value = true;
  loadError.value = "";
  try {
    const result = await loadCatalog({ type: filter.value, q: query.value.trim(), offset });
    if (requestId !== latestRequestId) return;
    rows.value = append ? [...rows.value, ...result.items] : result.items;
    hasMore.value = result.hasMore;
  } catch (error) {
    if (requestId === latestRequestId) loadError.value = error.message;
  } finally {
    if (requestId === latestRequestId) loading.value = false;
  }
}
async function openDetail(item) {
  const requestId = ++detailRequestId;
  selectedTitle.value = item;
  selectedDetail.value = null;
  detailError.value = "";
  detailLoading.value = true;
  try {
    selectedDetail.value = await loadCatalogDetail(item);
  } catch (error) {
    if (requestId === detailRequestId) detailError.value = error.message;
  } finally {
    if (requestId === detailRequestId) detailLoading.value = false;
  }
}
function closeDetail() {
  detailRequestId += 1;
  selectedTitle.value = null;
  selectedDetail.value = null;
  detailLoading.value = false;
  detailError.value = "";
}
watch([query, filter], () => refreshCatalog());
onMounted(refreshCatalog);
const emptyText = computed(() => loading.value ? "正在读取 PostgreSQL 片库…" : "数据库中还没有符合条件的影视条目。导入 TMDB 数据后会显示在这里。");
</script>

<template>
  <div class="view-stack">
    <section class="page-heading"><div><span class="eyebrow">CATALOG / LOCAL MIRROR</span><h1>影视片库</h1><p>按热度浏览海报，可按类型和片名筛选。</p></div><span class="data-pill"><i></i> {{ dashboard.catalog.size }} 本地数据</span></section>
    <section class="surface catalog-surface">
      <div class="catalog-toolbar">
        <div class="filter-tabs" role="group" aria-label="影视类型">
          <button v-for="item in filters" :key="item" :class="{ selected: filter === item }" :aria-pressed="filter === item" @click="filter = item">{{ item }}</button>
        </div>
        <label class="search-box"><span aria-hidden="true">⌕</span><input v-model="query" aria-label="搜索片名" placeholder="搜索本地片库（短词按开头匹配）" /></label>
      </div>
      <p v-if="loadError" class="error-banner" role="alert">{{ loadError }} <button @click="refreshCatalog">重试</button></p>
      <div v-if="loading && !rows.length" class="catalog-grid" aria-label="正在加载海报">
        <div v-for="skeleton in posterSkeletons" :key="skeleton" class="poster-skeleton"></div>
      </div>
      <div v-else-if="rows.length" class="catalog-grid" aria-label="影视海报墙">
        <CatalogPosterCard
          v-for="item in rows"
          :key="`${item.kind}-${item.id}`"
          :id="item.id"
          :title="item.title"
          :original-title="item.originalTitle"
          :kind="item.kind"
          :release-date="item.releaseDate"
          :poster-path="item.posterPath"
          @select="openDetail"
        />
      </div>
      <div v-else class="empty-state" role="status">{{ emptyText }}</div>
      <button v-if="hasMore" class="load-more" :disabled="loading" @click="refreshCatalog(true)">{{ loading ? "正在加载…" : "加载更多" }}</button>
      <div class="table-foot">当前显示 {{ rows.length.toLocaleString() }} 项 <span>PostgreSQL 共 {{ dashboard.catalog.titles.toLocaleString() }} 条</span></div>
    </section>
    <p class="preview-note"><span>i</span> 条目和搜索结果直接读取本机 PostgreSQL。</p>
    <CatalogDetailDialog
      v-if="selectedTitle"
      :selected="selectedTitle"
      :detail="selectedDetail"
      :loading="detailLoading"
      :error="detailError"
      @close="closeDetail"
    />
  </div>
</template>
