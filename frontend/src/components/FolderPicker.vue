<script setup>
import { ref } from "vue";
import { browseDirectories } from "../api";
const emit = defineEmits(["select"]);
const dialog = ref(null);
const current = ref("");
const parent = ref(null);
const roots = ref([]);
const directories = ref([]);
const selected = ref("");
const error = ref("");
const loading = ref(false);
async function open() {
  dialog.value?.showModal();
  await load("");
}
async function load(path) {
  loading.value = true;
  error.value = "";
  selected.value = "";
  try {
    const data = await browseDirectories(path);
    current.value = data.current;
    parent.value = data.parent;
    roots.value = data.roots;
    directories.value = data.directories;
  } catch (cause) {
    error.value = cause.message;
  } finally {
    loading.value = false;
  }
}
function choose() {
  if (selected.value) {
    emit("select", selected.value);
    dialog.value?.close();
  }
}
defineExpose({ open });
</script>
<template>
  <dialog ref="dialog" class="folder-picker" aria-labelledby="picker-title">
    <div class="picker-header">
      <div>
        <p class="eyebrow">已授权目录</p>
        <h2 id="picker-title">选择文件夹</h2>
      </div>
      <button class="icon-button" aria-label="关闭" @click="dialog?.close()">
        ×
      </button>
    </div>
    <div class="picker-toolbar">
      <button
        class="button secondary"
        :disabled="!parent || loading"
        @click="load(parent)"
      >
        上一级</button
      ><span class="picker-current">{{ current || "请选择一个授权路径" }}</span>
    </div>
    <div class="picker-list" aria-live="polite">
      <p v-if="loading" class="picker-empty">正在读取目录…</p>
      <template v-else-if="!current"
        ><button
          v-for="root in roots"
          :key="root"
          class="picker-entry"
          @click="load(root)"
        >
          <span>▰</span><strong>{{ root }}</strong
          ><span></span></button></template
      ><template v-else
        ><button
          v-for="folder in directories"
          :key="folder.path"
          class="picker-entry"
          :class="{ selected: selected === folder.path }"
          @click="selected = folder.path"
          @dblclick="load(folder.path)"
        >
          <span>▰</span><strong>{{ folder.name }}</strong
          ><span class="picker-entry-path">{{ folder.path }}</span
          ><span @click.stop="load(folder.path)">›</span>
        </button>
        <p v-if="!directories.length" class="picker-empty">
          此目录下没有子文件夹
        </p></template
      >
    </div>
    <p v-if="error" class="message error">{{ error }}</p>
    <div class="picker-actions">
      <button class="button secondary" @click="dialog?.close()">取消</button
      ><button class="button primary" :disabled="!current" @click="choose">
        选择此文件夹
      </button>
    </div>
  </dialog>
</template>
