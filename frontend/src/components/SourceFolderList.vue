<script setup>
import { ref } from "vue";
import FolderPicker from "./FolderPicker.vue";

const MAX_SOURCE_DIRECTORIES = 50;
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
});
const emit = defineEmits(["update:modelValue"]);
const picker = ref(null);
const error = ref("");

function addFolder(path) {
  if (props.modelValue.includes(path)) {
    error.value = "这个文件夹已添加";
    return;
  }
  if (props.modelValue.length >= MAX_SOURCE_DIRECTORIES) {
    error.value = `最多可添加 ${MAX_SOURCE_DIRECTORIES} 个文件夹`;
    return;
  }
  error.value = "";
  emit("update:modelValue", [...props.modelValue, path]);
}

function removeFolder(path) {
  error.value = "";
  emit(
    "update:modelValue",
    props.modelValue.filter((folder) => folder !== path),
  );
}
</script>

<template>
  <div class="source-folder-list">
    <div class="source-folder-heading">
      <div>
        <strong>待整理文件夹</strong>
        <small>可添加多个来源，整理到同一个片库目录</small>
      </div>
      <button
        class="button secondary"
        type="button"
        :disabled="modelValue.length >= MAX_SOURCE_DIRECTORIES"
        @click="picker?.open()"
      >
        添加文件夹
      </button>
    </div>
    <ul v-if="modelValue.length" class="source-folder-items">
      <li v-for="folder in modelValue" :key="folder">
        <span>{{ folder }}</span>
        <button
          class="button secondary source-folder-remove"
          type="button"
          :aria-label="`移除 ${folder}`"
          @click="removeFolder(folder)"
        >
          移除
        </button>
      </li>
    </ul>
    <p v-else class="source-folder-empty">请添加一个或多个已授权文件夹</p>
    <p v-if="error" class="message error" role="alert">{{ error }}</p>
    <FolderPicker ref="picker" @select="addFolder" />
  </div>
</template>
