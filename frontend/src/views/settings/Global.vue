<template>
  <errors
    v-if="error"
    :error-code="error.status"
  />
  <div
    v-else-if="!layoutStore.loading && settings !== null"
    class="row"
  >
    <div class="column">
      <form
        class="card"
        @submit.prevent="save"
      >
        <div class="card-title">
          <h2>{{ t("settings.globalSettings") }}</h2>
        </div>

        <div class="card-content">
          <p>
            <input
              v-model="settings.signup"
              type="checkbox"
            >
            {{ t("settings.allowSignup") }}
          </p>

          <p>
            <input
              v-model="settings.createUserDir"
              type="checkbox"
            >
            {{ t("settings.createUserDir") }}
          </p>

          <p>
            <input
              v-model="settings.hideLoginButton"
              type="checkbox"
            >
            {{ t("settings.hideLoginButton") }}
          </p>

          <p>
            <label class="small">{{ t("settings.userHomeBasePath") }}</label>
            <input
              v-model="settings.userHomeBasePath"
              class="input input--block"
              type="text"
            >
          </p>

          <p>
            <label for="minimumPasswordLength">{{
              t("settings.minimumPasswordLength")
            }}</label>
            <vue-number-input
              id="minimumPasswordLength"
              v-model.number="settings.minimumPasswordLength"
              controls
              :min="1"
            />
          </p>

          <h3>{{ t("settings.rules") }}</h3>
          <p class="small">
            {{ t("settings.globalRules") }}
          </p>
          <rules v-model:rules="settings.rules" />

          <div v-if="enableExec">
            <h3>{{ t("settings.executeOnShell") }}</h3>
            <p class="small">
              {{ t("settings.executeOnShellDescription") }}
            </p>
            <input
              v-model="shellValue"
              class="input input--block"
              type="text"
              placeholder="bash -c, cmd /c, ..."
            >
          </div>

          <h3>{{ t("settings.branding") }}</h3>

          <i18n-t
            keypath="settings.brandingHelp"
            tag="p"
            class="small"
            scope="global"
          >
            <a
              class="link"
              target="_blank"
              href="https://filebrowser.org/configuration.html#custom-branding"
            >{{ t("settings.documentation") }}</a>
          </i18n-t>

          <p>
            <input
              id="branding-links"
              v-model="settings.branding.disableExternal"
              type="checkbox"
            >
            {{ t("settings.disableExternalLinks") }}
          </p>

          <p>
            <input
              id="branding-used-disk"
              v-model="settings.branding.disableUsedPercentage"
              type="checkbox"
            >
            {{ t("settings.disableUsedDiskPercentage") }}
          </p>

          <p>
            <label for="theme">{{ t("settings.themes.title") }}</label>
            <themes
              id="theme"
              v-model:theme="settings.branding.theme"
              class="input input--block"
            />
          </p>

          <p>
            <label for="branding-name">{{ t("settings.instanceName") }}</label>
            <input
              id="branding-name"
              v-model="settings.branding.name"
              class="input input--block"
              type="text"
            >
          </p>

          <p>
            <label for="branding-files">{{
              t("settings.brandingDirectoryPath")
            }}</label>
            <input
              id="branding-files"
              v-model="settings.branding.files"
              class="input input--block"
              type="text"
            >
          </p>

          <h3>{{ t("settings.tusUploads") }}</h3>

          <p class="small">
            {{ t("settings.tusUploadsHelp") }}
          </p>

          <div class="tusConditionalSettings">
            <p>
              <label for="tus-chunkSize">{{
                t("settings.tusUploadsChunkSize")
              }}</label>
              <input
                id="tus-chunkSize"
                v-model="formattedChunkSize"
                class="input input--block"
                type="text"
              >
            </p>

            <p>
              <label for="tus-retryCount">{{
                t("settings.tusUploadsRetryCount")
              }}</label>
              <vue-number-input
                id="tus-retryCount"
                v-model.number="settings.tus.retryCount"
                controls
                :min="0"
              />
            </p>
          </div>
        </div>

        <div class="card-action">
          <input
            class="button button--flat"
            type="submit"
            :value="t('buttons.update')"
          >
        </div>
      </form>
    </div>

    <div class="column">
      <form
        class="card"
        @submit.prevent="save"
      >
        <div class="card-title">
          <h2>{{ t("settings.userDefaults") }}</h2>
        </div>

        <div class="card-content">
          <p class="small">
            {{ t("settings.defaultUserDescription") }}
          </p>

          <user-form
            v-model:user="settings.defaults"
            :is-new="false"
            :is-default="true"
          />
        </div>

        <div class="card-action">
          <input
            class="button button--flat"
            type="submit"
            :value="t('buttons.update')"
          >
        </div>
      </form>
    </div>

    <div class="column">
      <form
        v-if="enableExec"
        class="card"
        @submit.prevent="save"
      >
        <div class="card-title">
          <h2>{{ t("settings.commandRunner") }}</h2>
        </div>

        <div class="card-content">
          <i18n-t
            keypath="settings.commandRunnerHelp"
            tag="p"
            class="small"
            scope="global"
          >
            <code>FILE</code>
            <code>SCOPE</code>
            <a
              class="link"
              target="_blank"
              href="https://filebrowser.org/configuration.html#command-runner"
            >{{ t("settings.documentation") }}</a>
          </i18n-t>

          <div
            v-for="(command, key) in settings.commands"
            :key="key"
            class="collapsible"
          >
            <input
              :id="key"
              type="checkbox"
            >
            <label :for="key">
              <p>{{ capitalize(key) }}</p>
              <i class="material-icons">arrow_drop_down</i>
            </label>
            <div class="collapse">
              <textarea
                v-model.trim="commandObject[key]"
                class="input input--block input--textarea"
              />
            </div>
          </div>
        </div>

        <div class="card-action">
          <input
            class="button button--flat"
            type="submit"
            :value="t('buttons.update')"
          >
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { settings as api } from "@/api";
import { StatusError } from "@/api/utils";
import Rules from "@/components/settings/Rules.vue";
import Themes from "@/components/settings/Themes.vue";
import UserForm from "@/components/settings/UserForm.vue";
import { useLayoutStore } from "@/stores/layout";
import { enableExec } from "@/utils/constants";
import { getTheme, setTheme } from "@/utils/theme";
import Errors from "@/views/Errors.vue";
import { computed, inject, onBeforeUnmount, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const error = ref<StatusError | null>(null);
const originalSettings = ref<ISettings | null>(null);
const settings = ref<ISettings | null>(null);
const debounceTimeout = ref<number | null>(null);

const commandObject = ref<{
  [key: string]: string[] | string;
}>({});
const shellValue = ref<string>("");

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n();

const layoutStore = useLayoutStore();

const formattedChunkSize = computed({
  get() {
    return settings?.value?.tus?.chunkSize
      ? formatBytes(settings?.value?.tus?.chunkSize)
      : "";
  },
  set(value: string) {
    // Use debouncing to allow the user to type freely without
    // interruption by the formatter
    // Clear the previous timeout if it exists
    if (debounceTimeout.value) {
      clearTimeout(debounceTimeout.value);
    }

    // Set a new timeout to apply the format after a short delay
    debounceTimeout.value = window.setTimeout(() => {
      if (settings.value) settings.value.tus.chunkSize = parseBytes(value);
    }, 1500);
  },
});

// Define funcs
const capitalize = (name: string, where: string | RegExp = "_") => {
  if (where === "caps") where = /(?=[A-Z])/;
  const split = name.split(where);
  name = "";

  for (let i = 0; i < split.length; i++) {
    name += split[i].charAt(0).toUpperCase() + split[i].slice(1) + " ";
  }

  return name.slice(0, -1);
};

const save = async () => {
  if (settings.value === null) return false;
  const newSettings: ISettings = {
    ...settings.value,
    shell:
      settings.value?.shell
        .join(" ")
        .trim()
        .split(" ")
        .filter((s: string) => s !== "") ?? [],
    commands: {},
  };

  const keys = Object.keys(settings.value.commands) as Array<
    keyof SettingsCommand
  >;
  for (const key of keys) {
    // not sure if we can safely assume non-null
    const newValue = commandObject.value[key];
    if (!newValue) continue;

    if (Array.isArray(newValue)) {
      newSettings.commands[key] = newValue;
    } else if (key in commandObject.value) {
      newSettings.commands[key] = newValue
        .split("\n")
        .filter((cmd: string) => cmd !== "");
    }
  }
  newSettings.shell = shellValue.value
    .trim()
    .split(" ")
    .filter((s) => s !== "");

  if (newSettings.branding.theme !== getTheme()) {
    setTheme(newSettings.branding.theme);
  }

  try {
    await api.update(newSettings);
    $showSuccess(t("settings.settingsUpdated"));
  } catch (e: any) {
    $showError(e);
  }

  return true;
};
// Parse the user-friendly input (e.g., "20M" or "1T") to bytes
const parseBytes = (input: string) => {
  const regex = /^(\d+)(\.\d+)?(B|K|KB|M|MB|G|GB|T|TB)?$/i;
  const matches = input.match(regex);
  if (matches) {
    const size = parseFloat(matches[1].concat(matches[2] || ""));
    let unit: keyof SettingsUnit =
      matches[3].toUpperCase() as keyof SettingsUnit;
    if (!unit.endsWith("B")) {
      unit += "B";
    }
    const units: SettingsUnit = {
      KB: 1024,
      MB: 1024 ** 2,
      GB: 1024 ** 3,
      TB: 1024 ** 4,
    };
    return size * (units[unit as keyof SettingsUnit] || 1);
  } else {
    return 1024 ** 2;
  }
};
// Format the chunk size in bytes to user-friendly format
const formatBytes = (bytes: number) => {
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size}${units[unitIndex]}`;
};

// Define Hooks

onMounted(async () => {
  try {
    layoutStore.loading = true;
    const original: ISettings = await api.get();
    const newSettings: ISettings = { ...original, commands: {} };

    const keys = Object.keys(original.commands) as Array<keyof SettingsCommand>;
    for (const key of keys) {
      newSettings.commands[key] = original.commands[key];
      commandObject.value[key] = original.commands[key]!.join("\n");
    }

    originalSettings.value = original;
    settings.value = newSettings;
    shellValue.value = newSettings.shell.join(" ");
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
});

// Clear the debounce timeout when the component is destroyed
onBeforeUnmount(() => {
  if (debounceTimeout.value) {
    clearTimeout(debounceTimeout.value);
  }
});
</script>
