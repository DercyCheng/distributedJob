<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { useUserStore } from "./store/modules/user";

const userStore = useUserStore();

onMounted(async () => {
  if (localStorage.getItem("token")) {
    try {
      await userStore.fetchUserInfo();
    } catch (error) {
      // Handle token expiration or invalid token
      localStorage.removeItem("token");
    }
  }
});
</script>

<style>
/* Global styles are imported from assets/styles/main.scss */
html,
body {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: "Helvetica Neue", Helvetica, "PingFang SC", "Hiragino Sans GB",
    "Microsoft YaHei", Arial, sans-serif;
}

#app {
  height: 100vh;
}
</style>
