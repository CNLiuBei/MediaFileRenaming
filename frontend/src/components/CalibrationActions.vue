<script setup>
defineProps({
  busy: { type: Boolean, required: true },
  hasItems: { type: Boolean, required: true },
  targetCount: { type: Number, required: true },
  isCalibrating: { type: Boolean, required: true },
  progress: { type: String, required: true },
  message: { type: String, required: true },
});
defineEmits(["scan", "calibrate", "stop"]);
</script>

<template>
  <div class="actions preview-actions">
    <button class="button primary" :disabled="busy" @click="$emit('scan')">
      {{ busy ? "处理中…" : hasItems ? "更新预览" : "扫描并预览" }}
    </button>
    <button
      v-if="hasItems && !isCalibrating"
      class="button secondary"
      :disabled="busy || !targetCount"
      @click="$emit('calibrate')"
    >
      重新校准（{{ targetCount }} 个片名）
    </button>
    <button
      v-if="isCalibrating"
      class="button secondary"
      @click="$emit('stop')"
    >
      停止自动校准
    </button>
    <span class="message">{{ progress || message }}</span>
  </div>
</template>
