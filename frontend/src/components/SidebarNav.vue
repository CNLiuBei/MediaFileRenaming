<script setup>
import { navigation } from "../data/categories";
defineProps({
  activePage: { type: String, required: true },
  version: { type: String, required: true },
});
defineEmits(["navigate"]);
</script>
<template>
  <aside class="sidebar">
    <a
      class="sidebar-brand"
      href="#console"
      @click.prevent="$emit('navigate', 'console')"
      ><span class="brand-mark">影</span
      ><span
        ><strong>影像归档</strong><small>FNOS MEDIA LIBRARY</small></span
      ></a
    >
    <nav class="side-nav" aria-label="主导航">
      <button
        v-for="item in navigation"
        :key="item.id"
        class="nav-link"
        :class="{ active: activePage === item.id }"
        @click="$emit('navigate', item.id)"
      >
        <span class="nav-icon" aria-hidden="true">{{ item.icon }}</span
        ><span>{{ item.label }}</span>
      </button>
    </nav>
    <div class="sidebar-status">
      <span>版本</span><strong>{{ version }}</strong>
    </div>
  </aside>
</template>
