<template>
  <div class="tabs-view-container">
    <el-scrollbar class="tabs-scrollbar" ref="scrollbarRef">
      <div class="tabs-view">
        <router-link
          v-for="tag in visitedViews"
          :key="tag.path"
          :class="isActive(tag) ? 'active-tab' : 'tab'"
          :to="{ path: tag.path, query: tag.query }"
        >
          <span>{{ tag.meta.title }}</span>
          <el-icon
            class="tab-close"
            @click.prevent.stop="closeSelectedTag(tag)"
          >
            <Close />
          </el-icon>
        </router-link>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Close } from "@element-plus/icons-vue";
import type { RouteLocationNormalizedLoaded } from "vue-router";

const route = useRoute();
const router = useRouter();
const scrollbarRef = ref();

// Store visited views
const visitedViews = ref<RouteLocationNormalizedLoaded[]>([]);

// Add visited view
const addVisitedView = (view: RouteLocationNormalizedLoaded) => {
  if (!view.meta?.title) return;

  // Skip adding some routes
  if (view.path === "/login" || view.path === "/404") return;

  const exists = visitedViews.value.some((v) => v.path === view.path);
  if (!exists) {
    visitedViews.value.push(Object.assign({}, view));
  }
};

// Close selected tag
const closeSelectedTag = (view: RouteLocationNormalizedLoaded) => {
  const index = visitedViews.value.findIndex((v) => v.path === view.path);
  if (index > -1) {
    visitedViews.value.splice(index, 1);
  }

  // If closing active tag, navigate to the next available tag
  if (isActive(view)) {
    toLastView(visitedViews.value, view);
  }
};

// Check if current route matches the tag
const isActive = (view: RouteLocationNormalizedLoaded) => {
  return view.path === route.path;
};

// Navigate to another view after closing a tab
const toLastView = (
  visitedViews: RouteLocationNormalizedLoaded[],
  view: RouteLocationNormalizedLoaded
) => {
  const latestView = visitedViews.slice(-1)[0];

  if (latestView) {
    router.push(latestView.path);
  } else {
    if (view.name === "Dashboard") {
      router.push("/");
    } else {
      router.push("/");
    }
  }
};

// Watch route changes to add new tabs
watch(
  () => route.path,
  () => {
    addVisitedView(route);
  }
);

onMounted(() => {
  addVisitedView(route);
});
</script>

<style lang="scss" scoped>
.tabs-view-container {
  height: 34px;
  width: 100%;
  background: #fff;
  border-bottom: 1px solid #d8dce5;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.12), 0 0 3px 0 rgba(0, 0, 0, 0.04);

  .tabs-scrollbar {
    height: 34px;
    white-space: nowrap;

    .tabs-view {
      display: inline-block;
      padding-left: 10px;

      .tab {
        display: inline-block;
        height: 26px;
        line-height: 26px;
        border: 1px solid #d8dce5;
        color: #495060;
        background: #fff;
        padding: 0 8px;
        font-size: 12px;
        margin-left: 5px;
        margin-top: 4px;
        border-radius: 3px;
        text-decoration: none;

        &:first-of-type {
          margin-left: 0;
        }

        .tab-close {
          margin-left: 5px;
          color: #666;
          vertical-align: middle;
          font-size: 12px;
          transition: color 0.3s;

          &:hover {
            color: #f56c6c;
          }
        }
      }

      .active-tab {
        display: inline-block;
        height: 26px;
        line-height: 26px;
        border: 1px solid var(--primary-color);
        color: #fff;
        background-color: var(--primary-color);
        padding: 0 8px;
        font-size: 12px;
        margin-left: 5px;
        margin-top: 4px;
        border-radius: 3px;
        text-decoration: none;

        &:first-of-type {
          margin-left: 0;
        }

        .tab-close {
          margin-left: 5px;
          color: #fff;
          vertical-align: middle;
          font-size: 12px;
          transition: color 0.3s;

          &:hover {
            color: #f2f2f2;
          }
        }
      }
    }
  }
}
</style>
