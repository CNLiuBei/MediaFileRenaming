<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { getStatus } from "./api";
import SidebarNav from "./components/SidebarNav.vue";
import ConsolePage from "./pages/ConsolePage.vue";
import SettingsPage from "./pages/SettingsPage.vue";
import LogsPage from "./pages/LogsPage.vue";
import { appVersion } from "./data/appInfo";

const activePage = ref("console");
const isOnline = ref(false);
const loadError = ref("");
const config = reactive({
  source_dir: "",
  destination_dir: "",
  watch_enabled: false,
  interval_minutes: 15,
  tmdb_key_configured: false,
  tmdb_proxy_url: "",
  offline_catalog: { available: false, titles: 0, aliases: 0, episodes: 0 },
});
const authorizedRoots = ref([]);
const pageTitle = computed(
  () =>
    ({ console: "控制台", settings: "设置", logs: "日志" })[activePage.value],
);
async function loadStatus() {
  try {
    const result = await getStatus();
    Object.assign(config, result.config);
    authorizedRoots.value = result.authorized_roots;
    isOnline.value = true;
    loadError.value = "";
  } catch (error) {
    isOnline.value = false;
    loadError.value = error.message;
  }
}
function updateConfig(value) {
  Object.assign(config, value);
}
onMounted(loadStatus);
</script>

<template>
  <div class="app-shell">
    <SidebarNav
      :active-page="activePage"
      :version="appVersion"
      @navigate="activePage = $event"
    />
    <main class="workspace">
      <header class="topbar">
        <div>
          <p class="eyebrow">影视文件整理</p>
          <h1>{{ pageTitle }}</h1>
        </div>
        <span class="status-pill" :class="{ offline: !isOnline }">{{
          isOnline ? "服务运行中" : "服务未连接"
        }}</span>
      </header>
      <p v-if="loadError" class="page-error" role="alert">{{ loadError }}</p>
      <KeepAlive>
        <ConsolePage v-if="activePage === 'console'" :config="config" />
        <SettingsPage
          v-else-if="activePage === 'settings'"
          :config="config"
          :authorized-roots="authorizedRoots"
          @saved="updateConfig"
          @catalog-imported="loadStatus"
        />
        <LogsPage v-else />
      </KeepAlive>
      <footer>
        文件整理在本机处理。离线片库未命中时，在线搜索会向 TMDB 发送片名。本产品使用 TMDB
        API，但未获 TMDB 认可或认证。
      </footer>
    </main>
  </div>
</template>
