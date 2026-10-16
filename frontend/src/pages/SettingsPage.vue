<script setup>
import { reactive, ref, watch } from "vue";
import { saveConfig } from "../api";
import { intervals } from "../data/categories";
import FolderPicker from "../components/FolderPicker.vue";
import OfflineCatalogPanel from "../components/OfflineCatalogPanel.vue";
import SyncStationPanel from "../components/SyncStationPanel.vue";

const props = defineProps({
  config: { type: Object, required: true },
  authorizedRoots: { type: Array, required: true },
});
const emit = defineEmits(["saved", "catalog-imported"]);
const picker = ref(null);
const folderTarget = ref("source_dir");
const message = ref("");
const error = ref("");
const busy = ref(false);
const draft = reactive({ ...props.config, tmdb_api_key: "" });
watch(
  () => props.config,
  (value) => Object.assign(draft, value),
  { deep: true },
);
function chooseFolder(target) {
  folderTarget.value = target;
  picker.value?.open();
}
function setFolder(path) {
  draft[folderTarget.value] = path;
}
async function save() {
  busy.value = true;
  message.value = "";
  error.value = "";
  try {
    const payload = {
      source_dir: draft.source_dir,
      destination_dir: draft.destination_dir,
      watch_enabled: draft.watch_enabled,
      interval_minutes: draft.interval_minutes,
      tmdb_proxy_url: draft.tmdb_proxy_url || "",
      tmdb_sync_station_url: draft.tmdb_sync_station_url || "",
    };
    if (draft.tmdb_api_key.trim())
      payload.tmdb_api_key = draft.tmdb_api_key.trim();
    const result = await saveConfig(payload);
    emit("saved", result.config);
    draft.tmdb_key_configured = result.config.tmdb_key_configured;
    draft.tmdb_api_key = "";
    message.value = "设置已保存。";
  } catch (cause) {
    error.value = cause.message;
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <section class="panel settings-panel">
    <div class="section-heading">
      <div>
        <p class="eyebrow">目录设置</p>
        <h2>选择整理范围</h2>
      </div>
      <span class="step-dot">01</span>
    </div>
    <div class="field-grid">
      <label
        >待整理目录
        <div class="directory-field">
          <input
            v-model="draft.source_dir"
            type="text"
            placeholder="请选择已授权目录"
          /><button
            class="button secondary"
            @click="chooseFolder('source_dir')"
          >
            选择文件夹
          </button>
        </div></label
      ><label
        >整理到目录
        <div class="directory-field">
          <input
            v-model="draft.destination_dir"
            type="text"
            placeholder="请选择已授权目录"
          /><button
            class="button secondary"
            @click="chooseFolder('destination_dir')"
          >
            选择文件夹
          </button>
        </div></label
      >
    </div>
    <p class="hint">
      目录必须位于 fnOS 授权路径中。请先在应用设置中授权对应目录。
    </p>
    <label class="tmdb-key-field"
      >TMDB API Key（可选）<input
        v-model="draft.tmdb_api_key"
        type="password"
        autocomplete="new-password"
        :placeholder="
          draft.tmdb_key_configured
            ? '已配置；留空保留现有 Key'
            : '用于搜索并校准影视信息'
        "
    /></label>
    <p class="hint">
      填写 TMDB 账户 API 设置中的 v3 API
      Key。搜索候选会显示名称、年份、语言、地区和简介；选择后会建议二级分类。
    </p>
    <label class="tmdb-key-field"
      >TMDB API 代理地址（可选）<input
        v-model="draft.tmdb_proxy_url"
        type="url"
        autocomplete="off"
        placeholder="https://tmdb.liubei.org"
    /></label>
    <p class="hint">
      填写 NAS 可访问的 TMDB API 代理域名，例如
      https://tmdb.liubei.org；也可粘贴完整的 TMDB API
      链接。留空表示直连。此处不是 HTTP 出口代理设置。TMDB 校准请求限速为 0.8
      QPS（至少间隔 1.25 秒）。
    </p>
    <SyncStationPanel
      v-model="draft.tmdb_sync_station_url"
      :status="props.config.tmdb_sync_station || {}"
    />
    <div class="allowed-roots">
      <span>已授权路径</span>
      <div>
        <span v-if="!authorizedRoots.length">尚未读取到授权路径</span
        ><span v-for="root in authorizedRoots" :key="root" class="root-path">{{
          root
        }}</span>
      </div>
    </div>
    <p v-if="error" class="message error" role="alert">{{ error }}</p>
    <div class="actions">
      <button class="button primary" :disabled="busy" @click="save">
        {{ busy ? "保存中…" : "保存目录设置" }}</button
      ><span class="message" role="status">{{ message }}</span>
    </div>
  </section>
  <section class="panel automation-panel">
    <div class="section-heading">
      <div>
        <p class="eyebrow">自动归档</p>
        <h2>新文件自动整理</h2>
      </div>
      <span class="step-dot">02</span>
    </div>
    <p class="muted">
      开启后，应用按设定间隔扫描待整理目录并移动可识别的文件。自动整理使用文件名识别结果。
    </p>
    <div class="automation-controls">
      <label class="toggle-row"
        ><input v-model="draft.watch_enabled" type="checkbox" /><span
          class="toggle"
        ></span
        ><span
          ><strong>启用自动整理</strong
          ><small>关闭时仅在手动执行后移动文件</small></span
        ></label
      ><label class="interval-field"
        >扫描间隔<select v-model.number="draft.interval_minutes">
          <option v-for="minutes in intervals" :key="minutes" :value="minutes">
            {{
              minutes < 60
                ? `每 ${minutes} 分钟`
                : minutes === 60
                  ? "每小时"
                  : minutes === 1440
                    ? "每天"
                    : `每 ${minutes / 60} 小时`
            }}
          </option>
        </select></label
      >
    </div>
    <div class="actions">
      <button class="button secondary" :disabled="busy" @click="save">
        保存自动整理设置</button
      ><span class="message">{{ message }}</span>
    </div>
  </section>
  <OfflineCatalogPanel :catalog="config.offline_catalog" @imported="emit('catalog-imported', $event)" />
  <FolderPicker ref="picker" @select="setFolder" />
</template>
