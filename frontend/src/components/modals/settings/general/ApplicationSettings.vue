<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhPalette, PhMoon, PhTranslate, PhPower, PhArchiveTray } from '@phosphor-icons/vue';
import { SettingGroup, SettingWithToggle, SettingWithSelect } from '@/components/settings';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

function updateSetting(key: keyof SettingsData, value: any) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });
}
</script>

<template>
  <SettingGroup :icon="PhPalette" :title="t('setting.general.application')">
    <SettingWithToggle
      :icon="PhPower"
      :title="t('setting.general.startupOnBoot')"
      :description="t('setting.general.startupOnBootDesc')"
      :model-value="settings.startup_on_boot"
      @update:model-value="updateSetting('startup_on_boot', $event)"
    />

    <SettingWithToggle
      :icon="PhArchiveTray"
      :title="t('setting.general.closeToTray')"
      :description="t('setting.general.closeToTrayDesc')"
      :model-value="settings.close_to_tray"
      @update:model-value="updateSetting('close_to_tray', $event)"
    />

    <SettingWithSelect
      :icon="PhMoon"
      :title="t('setting.general.theme')"
      :description="t('setting.general.themeDesc')"
      :model-value="settings.theme"
      :options="[
        { value: 'light', label: t('setting.general.light') },
        { value: 'dark', label: t('setting.general.dark') },
        { value: 'auto', label: t('setting.general.auto') },
      ]"
      width="md"
      @update:model-value="updateSetting('theme', $event)"
    />

    <SettingWithSelect
      :icon="PhTranslate"
      :title="t('setting.general.language')"
      :description="t('setting.general.languageDesc')"
      :model-value="settings.language"
      :options="[
        { value: 'en-US', label: t('common.language.english') },
        { value: 'zh-CN', label: t('common.language.chinese') },
      ]"
      width="md"
      @update:model-value="updateSetting('language', $event)"
    />
  </SettingGroup>
</template>

<style scoped></style>
