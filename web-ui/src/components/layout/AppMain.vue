<template>
  <section class="app-main">
    <router-view v-slot="{ Component }">
      <transition name="fade-transform" mode="out-in">
        <keep-alive :include="cachedViews">
          <component :is="Component" :key="key" />
        </keep-alive>
      </transition>
    </router-view>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const key = computed(() => route.path);

// List of components that should be cached
const cachedViews = ref([
  "Dashboard",
  "TaskList",
  "RecordList",
  "DepartmentList",
  "UserList",
  "RoleList",
]);
</script>

<style lang="scss" scoped>
.app-main {
  padding: 20px;
  height: calc(100vh - 84px); // 50px header + 34px tabs
  overflow-y: auto;
  box-sizing: border-box;
}

/* Transition effects */
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
