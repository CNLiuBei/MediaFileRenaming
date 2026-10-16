<script setup>
import { computed } from "vue";
import PlanRow from "./PlanRow.vue";
import PosterPreview from "./PosterPreview.vue";

const props = defineProps({
  group: { type: Object, required: true },
  imageBaseUrl: { type: String, required: true },
});
defineEmits(["change"]);
const posterPath = computed(
  () =>
    props.group.poster_path ||
    props.group.items.find((item) => item.poster_path)?.poster_path ||
    "",
);
</script>

<template>
  <section class="series-group">
    <header class="series-group-header">
      <PosterPreview
        :poster-path="posterPath"
        :image-base-url="imageBaseUrl"
        :alt="`${group.title} 海报`"
      />
      <div>
        <p class="eyebrow">电视剧 · {{ group.items.length }} 个视频文件</p>
        <h3>{{ group.title }}</h3>
        <p class="series-group-hint">
          任意一集完成 TMDB 校准后，片名、分类和海报会应用到本组所有集数。
        </p>
      </div>
    </header>
    <PlanRow
      v-for="item in group.items"
      :key="item.source"
      :item="item"
      :image-base-url="imageBaseUrl"
      :show-poster="false"
      @change="$emit('change', $event)"
    />
  </section>
</template>
