<template>
  <div class="dashboard">
    <header-bar
      show-menu
      show-logo
    />

    <div id="nav">
      <div class="wrapper">
        <ul>
          <router-link to="/settings/profile">
            <li :class="{ active: $route.path === '/settings/profile' }">
              {{ t("settings.profileSettings") }}
            </li>
          </router-link>
          <router-link
            v-if="user?.perm.share"
            to="/settings/shares"
          >
            <li :class="{ active: $route.path === '/settings/shares' }">
              {{ t("settings.shareManagement") }}
            </li>
          </router-link>
          <router-link
            v-if="user?.perm.admin"
            to="/settings/global"
          >
            <li :class="{ active: $route.path === '/settings/global' }">
              {{ t("settings.globalSettings") }}
            </li>
          </router-link>
          <router-link
            v-if="user?.perm.admin"
            to="/settings/users"
          >
            <li
              :class="{
                active:
                  $route.path === '/settings/users' || $route.name === 'User',
              }"
            >
              {{ t("settings.userManagement") }}
            </li>
          </router-link>
        </ul>
      </div>
    </div>

    <div v-if="loading">
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1" />
          <div class="bounce2" />
          <div class="bounce3" />
        </div>
        <span>{{ t("files.loading") }}</span>
      </h2>
    </div>

    <router-view />
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const authStore = useAuthStore();
const layoutStore = useLayoutStore();

const user = computed(() => authStore.user);
const loading = computed(() => layoutStore.loading);
</script>
