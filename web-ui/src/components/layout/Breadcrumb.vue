<template>
  <el-breadcrumb class="app-breadcrumb" separator="/">
    <transition-group name="breadcrumb">
      <el-breadcrumb-item v-for="(item, index) in breadcrumbs" :key="item.path">
        <span v-if="index === breadcrumbs.length - 1" class="no-redirect">{{
          item.meta.title
        }}</span>
        <a v-else @click.prevent="handleLink(item)">{{ item.meta.title }}</a>
      </el-breadcrumb-item>
    </transition-group>
  </el-breadcrumb>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute, useRouter, RouteLocationMatched } from "vue-router";

const route = useRoute();
const router = useRouter();

const breadcrumbs = ref<RouteLocationMatched[]>([]);

const getBreadcrumbs = () => {
  let matched = route.matched.filter((item) => item.meta && item.meta.title);
  const first = matched[0];
  if (first && first.path !== "/dashboard") {
    matched = [{ path: "/dashboard", meta: { title: "首页" } } as any].concat(
      matched
    );
  }
  breadcrumbs.value = matched;
};

const handleLink = (item: RouteLocationMatched) => {
  router.push(item.path);
};

watch(
  () => route.path,
  () => getBreadcrumbs(),
  { immediate: true }
);
</script>

<style lang="scss" scoped>
.app-breadcrumb {
  display: inline-block;
  font-size: 14px;
  line-height: 50px;
  margin-left: 8px;

  .no-redirect {
    color: #97a8be;
    cursor: text;
  }
}

.breadcrumb-enter-active,
.breadcrumb-leave-active {
  transition: all 0.5s;
}

.breadcrumb-enter-from,
.breadcrumb-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.breadcrumb-leave-active {
  position: absolute;
}
</style>
