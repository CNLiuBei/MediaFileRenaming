<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { loadDashboard, loadDatabaseHealth, loadSyncStatus, startSync } from "./api";
import SidebarNav from "./components/SidebarNav.vue";
import ActivityView from "./views/ActivityView.vue";
import CatalogView from "./views/CatalogView.vue";
import OverviewView from "./views/OverviewView.vue";
import SettingsView from "./views/SettingsView.vue";

const page = ref("overview");
const dashboard = ref(null);
const database = ref(null);
const error = ref("");
const syncing = computed(() => dashboard.value?.sync.state === "running");
const title = computed(() => ({ overview: "总览", catalog: "影视片库", activity: "同步记录", settings: "设置" })[page.value]);
let syncPoll = null;

async function refresh() {
  try {
    const [nextDashboard, health] = await Promise.all([loadDashboard(), loadDatabaseHealth()]);
    dashboard.value = nextDashboard;
    database.value = health;
    error.value = "";
  } catch (cause) {
    error.value = `${cause.message}。请确认本机 PostgreSQL 和 Go 服务已启动。`;
  }
}
async function refreshSync() {
  try {
    const sync = await loadSyncStatus();
    if (dashboard.value) dashboard.value = { ...dashboard.value, sync };
    error.value = "";
  } catch (cause) {
    error.value = `${cause.message}。请确认本机 PostgreSQL 和 Go 服务已启动。`;
  }
}
async function requestSync() {
  try {
    await startSync();
    await refreshSync();
  } catch (cause) {
    if (cause.message.includes("已有同步任务正在运行")) {
      await refreshSync();
      return;
    }
    error.value = cause.message;
  }
}
watch(syncing, (isRunning, wasRunning) => {
  if (syncPoll) window.clearInterval(syncPoll);
  syncPoll = isRunning ? window.setInterval(refreshSync, 1500) : null;
  if (!isRunning && wasRunning) refresh();
});
onMounted(refresh);
onUnmounted(() => { if (syncPoll) window.clearInterval(syncPoll); });
</script>

<template>
  <div class="app-shell">
    <SidebarNav :active="page" @navigate="page = $event" />
    <main class="main-area">
      <header class="topbar"><div class="mobile-brand"><span class="brand-mark"><i></i><i></i><i></i></span>REEL</div><div class="crumb"><span>TMDB CATALOG STATION</span><b>/</b>{{ title }}</div><div class="top-actions"><span class="connection-pill"><i></i> {{ database ? 'PostgreSQL 已连接' : '连接数据库中' }}</span><button class="avatar" aria-label="本地用户">IF</button></div></header>
      <div v-if="error" class="error-banner" role="alert">{{ error }} <button @click="refresh">重试</button></div>
      <div v-if="dashboard" class="page-content">
        <OverviewView v-if="page === 'overview'" :dashboard="dashboard" :busy="syncing" @sync="requestSync" @all-changes="page = 'activity'" />
        <CatalogView v-else-if="page === 'catalog'" :dashboard="dashboard" />
        <ActivityView v-else-if="page === 'activity'" :dashboard="dashboard" />
        <SettingsView v-else :database="database" :dashboard="dashboard" />
      </div>
      <div v-else class="loading-state">正在连接本地 PostgreSQL 服务…</div>
      <footer class="app-footer"><span>REEL / PERSONAL CATALOG</span><span>Uses TMDB API; not endorsed or certified by TMDB</span></footer>
    </main>
  </div>
</template>
