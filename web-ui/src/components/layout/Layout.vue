<template>
  <div class="layout-container">
    <el-container>
      <el-aside width="220px" class="sidebar-container">
        <div class="logo">
          <h1>DistributedJob</h1>
        </div>
        <el-menu
          :default-active="activeMenu"
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
          :collapse="isCollapse"
          router
        >
          <sidebar-item
            v-for="route in routes"
            :key="route.path"
            :item="route"
            :base-path="route.path"
          />
        </el-menu>
      </el-aside>
      <el-container>
        <el-header height="50px" class="header">
          <div class="header-left">
            <el-icon @click="toggleCollapse" class="collapse-btn">
              <component :is="isCollapse ? 'Expand' : 'Fold'" />
            </el-icon>
            <breadcrumb />
          </div>
          <div class="header-right">
            <el-dropdown trigger="click" @command="handleCommand">
              <div class="user-info">
                <el-avatar :size="32" icon="el-icon-user" />
                <span class="username">{{ userStore.userInfo.name }}</span>
                <el-icon><arrow-down /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile"
                    >个人信息</el-dropdown-item
                  >
                  <el-dropdown-item command="logout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>
        <el-main>
          <tabs-view />
          <app-main />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import { ArrowDown, Expand, Fold } from "@element-plus/icons-vue";
import { useUserStore } from "@/store/modules/user";
import { logout } from "@/api/auth";
import SidebarItem from "./SidebarItem.vue";
import Breadcrumb from "./Breadcrumb.vue";
import TabsView from "./TabsView.vue";
import AppMain from "./AppMain.vue";

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();

const isCollapse = ref(false);
const routes = computed(() => {
  return router.options.routes.filter((route) => {
    return route.meta && !route.meta.hidden;
  });
});

const activeMenu = computed(() => {
  const { meta, path } = route;
  if (meta.activeMenu) {
    return meta.activeMenu;
  }
  return path;
});

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value;
};

const handleCommand = async (command: string) => {
  if (command === "logout") {
    try {
      await logout();
      localStorage.removeItem("token");
      router.push("/login");
    } catch (error) {
      console.error("Logout failed", error);
    }
  } else if (command === "profile") {
    // Navigate to profile page
  }
};
</script>

<style scoped lang="scss">
.layout-container {
  height: 100%;

  .sidebar-container {
    background-color: #304156;
    transition: width 0.28s;
    overflow-y: auto;

    .logo {
      height: 50px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      background-color: #2b2f3a;

      h1 {
        margin: 0;
        font-size: 18px;
        font-weight: 600;
      }
    }
  }

  .header {
    background-color: #fff;
    border-bottom: 1px solid #e6e6e6;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 15px;

    .header-left {
      display: flex;
      align-items: center;

      .collapse-btn {
        margin-right: 15px;
        font-size: 20px;
        cursor: pointer;
      }
    }

    .header-right {
      .user-info {
        display: flex;
        align-items: center;
        cursor: pointer;

        .username {
          margin: 0 5px;
        }
      }
    }
  }

  .el-main {
    padding: 0;
    height: calc(100vh - 50px);
    overflow: hidden;
    position: relative;
  }
}
</style>
