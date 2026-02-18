<template>
  <div>
    <p v-if="!isDefault && props.user !== null">
      <label for="username">{{ t("settings.username") }}</label>
      <input
        id="username"
        v-model="user.username"
        class="input input--block"
        type="text"
      >
    </p>

    <p v-if="!isDefault">
      <label for="password">{{ t("settings.password") }}</label>
      <input
        id="password"
        v-model="user.password"
        class="input input--block"
        type="password"
        :placeholder="passwordPlaceholder"
      >
    </p>

    <p>
      <label for="scope">{{ t("settings.scope") }}</label>
      <input
        id="scope"
        v-model="user.scope"
        :disabled="createUserDirData ?? false"
        :placeholder="scopePlaceholder"
        class="input input--block"
        type="text"
      >
    </p>
    <p
      v-if="displayHomeDirectoryCheckbox"
      class="small"
    >
      <input
        v-model="createUserDirData"
        type="checkbox"
      >
      {{ t("settings.createUserHomeDirectory") }}
    </p>

    <p>
      <label for="locale">{{ t("settings.language") }}</label>
      <languages
        id="locale"
        v-model:locale="user.locale"
        class="input input--block"
      />
    </p>

    <p v-if="!isDefault && user.perm">
      <input
        v-model="user.lockPassword"
        type="checkbox"
        :disabled="user.perm.admin"
      >
      {{ t("settings.lockPassword") }}
    </p>

    <permissions v-model:perm="user.perm" />
    <commands
      v-if="enableExec"
      v-model:commands="user.commands"
    />

    <div v-if="!isDefault">
      <h3>{{ t("settings.rules") }}</h3>
      <p class="small">
        {{ t("settings.rulesHelp") }}
      </p>
      <rules v-model:rules="user.rules" />
    </div>
  </div>
</template>

<script setup lang="ts">
import Languages from "./Languages.vue";
import Rules from "./Rules.vue";
import Permissions from "./Permissions.vue";
import Commands from "./Commands.vue";
import { enableExec } from "@/utils/constants";
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const createUserDirData = ref<boolean | null>(null);
const originalUserScope = ref<string | null>(null);

const props = defineProps<{
  user: IUserForm;
  isNew: boolean;
  isDefault: boolean;
  createUserDir?: boolean;
}>();

onMounted(() => {
  if (props.user.scope) {
    originalUserScope.value = props.user.scope;
    createUserDirData.value = props.createUserDir;
  }
});

const passwordPlaceholder = computed(() =>
  props.isNew ? "" : t("settings.avoidChanges")
);
const scopePlaceholder = computed(() =>
  createUserDirData.value ? t("settings.userScopeGenerationPlaceholder") : ""
);
const displayHomeDirectoryCheckbox = computed(
  () => props.isNew && createUserDirData.value
);

watch(
  () => props.user,
  () => {
    if (!props.user?.perm?.admin) return;
    props.user.lockPassword = false;
  }
);

watch(createUserDirData, () => {
  if (props.user?.scope) {
    props.user.scope = createUserDirData.value
      ? ""
      : (originalUserScope.value ?? "");
  }
});
</script>
