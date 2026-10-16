<script setup>
defineProps({ readyCount: { type: Number, required: true }, busy: { type: Boolean, required: true } });
defineEmits(["execute"]);
</script>

<template>
  <div class="actions apply-actions">
    <button class="button danger" :disabled="busy || !readyCount" @click="$emit('execute')">
      执行整理（{{ readyCount }}）
    </button>
    <span class="hint-inline">不会覆盖同名文件；无法识别的文件可手动校准。</span>
  </div>
</template>
