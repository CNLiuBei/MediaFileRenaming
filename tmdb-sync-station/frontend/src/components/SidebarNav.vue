<script setup>
defineProps({ active: { type: String, required: true } });
defineEmits(["navigate"]);

const items = [
  { id: "overview", label: "总览", icon: "◫" },
  { id: "catalog", label: "片库", icon: "▤" },
  { id: "activity", label: "同步记录", icon: "↻" },
  { id: "settings", label: "设置", icon: "⚙" },
];
</script>

<template>
  <aside class="sidebar">
    <div class="brand-lockup">
      <div class="brand-mark"><span></span><span></span><span></span></div>
      <div><strong>REEL</strong><small>TMDB CATALOG STATION</small></div>
    </div>
    <div class="nav-caption">数据工作区</div>
    <nav aria-label="主导航">
      <button v-for="item in items" :key="item.id" class="nav-item" :class="{ active: active === item.id }" @click="$emit('navigate', item.id)">
        <span class="nav-icon">{{ item.icon }}</span><span>{{ item.label }}</span><i v-if="active === item.id"></i>
      </button>
    </nav>
    <div class="sidebar-bottom">
      <div class="local-badge"><span class="pulse-dot"></span><span>本地服务</span><b>在线</b></div>
      <p>私有片库<br />数据保存在这台 NAS</p>
      <small>PREVIEW BUILD 0.1.0</small>
    </div>
  </aside>
</template>
