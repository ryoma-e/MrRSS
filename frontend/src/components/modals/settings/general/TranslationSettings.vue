<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhGlobe,
  PhArticle,
  PhPackage,
  PhKey,
  PhLink,
  PhRobot,
  PhChartLine,
  PhArrowCounterClockwise,
} from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

// AI usage tracking
const aiUsage = ref<{
  usage: number;
  limit: number;
  limit_reached: boolean;
}>({
  usage: 0,
  limit: 0,
  limit_reached: false,
});

async function fetchAIUsage() {
  try {
    const response = await fetch('/api/ai-usage');
    if (response.ok) {
      aiUsage.value = await response.json();
    }
  } catch (e) {
    console.error('Failed to fetch AI usage:', e);
  }
}

async function resetAIUsage() {
  if (!window.confirm(t('aiUsageResetConfirm'))) {
    return;
  }
  try {
    const response = await fetch('/api/ai-usage/reset', { method: 'POST' });
    if (response.ok) {
      await fetchAIUsage();
      // Reset the local settings value as well
      emit('update:settings', {
        ...props.settings,
        ai_usage_tokens: '0',
      });
      window.showToast(t('aiUsageResetSuccess'), 'success');
    }
  } catch (e) {
    console.error('Failed to reset AI usage:', e);
    window.showToast(t('aiUsageResetError'), 'error');
  }
}

// Calculate usage percentage
function getUsagePercentage(): number {
  if (aiUsage.value.limit === 0) return 0;
  return Math.min(100, (aiUsage.value.usage / aiUsage.value.limit) * 100);
}

onMounted(() => {
  fetchAIUsage();
});

// Refresh AI usage when provider changes to AI
watch(
  () => props.settings.translation_provider,
  (newProvider) => {
    if (newProvider === 'ai') {
      fetchAIUsage();
    }
  }
);
</script>

<template>
  <div class="setting-group">
    <label
      class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
    >
      <PhGlobe :size="14" class="sm:w-4 sm:h-4" />
      {{ t('translation') }}
    </label>
    <div class="setting-item mb-2 sm:mb-4">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhArticle :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('enableTranslation') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('enableTranslationDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="props.settings.translation_enabled"
        type="checkbox"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              translation_enabled: (e.target as HTMLInputElement).checked,
            })
        "
      />
    </div>

    <div
      v-if="props.settings.translation_enabled"
      class="ml-2 sm:ml-4 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4"
    >
      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhPackage :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('translationProvider') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('translationProviderDesc') || 'Choose the translation service to use' }}
            </div>
          </div>
        </div>
        <select
          :value="props.settings.translation_provider"
          class="input-field w-32 sm:w-48 text-xs sm:text-sm"
          @change="
            (e) =>
              emit('update:settings', {
                ...props.settings,
                translation_provider: (e.target as HTMLSelectElement).value,
              })
          "
        >
          <option value="google">{{ t('googleTranslate') }}</option>
          <option value="deepl">{{ t('deeplApi') }}</option>
          <option value="baidu">{{ t('baiduTranslate') }}</option>
          <option value="ai">{{ t('aiTranslation') }}</option>
        </select>
      </div>

      <!-- Google Translate Endpoint -->
      <div v-if="props.settings.translation_provider === 'google'" class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhLink :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('googleTranslateEndpoint') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('googleTranslateEndpointDesc') }}
            </div>
          </div>
        </div>
        <select
          :value="props.settings.google_translate_endpoint"
          class="input-field w-32 sm:w-48 text-xs sm:text-sm"
          @change="
            (e) =>
              emit('update:settings', {
                ...props.settings,
                google_translate_endpoint: (e.target as HTMLSelectElement).value,
              })
          "
        >
          <option value="translate.googleapis.com">
            {{ t('googleTranslateEndpointDefault') }}
          </option>
          <option value="clients5.google.com">{{ t('googleTranslateEndpointAlternate') }}</option>
        </select>
      </div>

      <!-- DeepL API Key -->
      <div v-if="props.settings.translation_provider === 'deepl'" class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhKey :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">
              {{ t('deeplApiKey') }}
              <span v-if="!props.settings.deepl_endpoint?.trim()" class="text-red-500">*</span>
            </div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('deeplApiKeyDesc') || 'Enter your DeepL API key' }}
            </div>
          </div>
        </div>
        <input
          :value="props.settings.deepl_api_key"
          type="password"
          :placeholder="t('deeplApiKeyPlaceholder')"
          :class="[
            'input-field w-32 sm:w-48 text-xs sm:text-sm',
            props.settings.translation_enabled &&
            props.settings.translation_provider === 'deepl' &&
            !props.settings.deepl_api_key?.trim() &&
            !props.settings.deepl_endpoint?.trim()
              ? 'border-red-500'
              : '',
          ]"
          @input="
            (e) =>
              emit('update:settings', {
                ...props.settings,
                deepl_api_key: (e.target as HTMLInputElement).value,
              })
          "
        />
      </div>

      <!-- DeepL Custom Endpoint (deeplx) -->
      <div v-if="props.settings.translation_provider === 'deepl'" class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhLink :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('deeplEndpoint') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('deeplEndpointDesc') }}
            </div>
          </div>
        </div>
        <input
          :value="props.settings.deepl_endpoint"
          type="text"
          :placeholder="t('deeplEndpointPlaceholder')"
          class="input-field w-32 sm:w-48 text-xs sm:text-sm"
          @input="
            (e) =>
              emit('update:settings', {
                ...props.settings,
                deepl_endpoint: (e.target as HTMLInputElement).value,
              })
          "
        />
      </div>

      <!-- Baidu Translate Settings -->
      <template v-if="props.settings.translation_provider === 'baidu'">
        <div class="sub-setting-item">
          <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhKey :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-sm">
                {{ t('baiduAppId') }} <span class="text-red-500">*</span>
              </div>
              <div class="text-xs text-text-secondary hidden sm:block">
                {{ t('baiduAppIdDesc') }}
              </div>
            </div>
          </div>
          <input
            :value="props.settings.baidu_app_id"
            type="text"
            :placeholder="t('baiduAppIdPlaceholder')"
            :class="[
              'input-field w-32 sm:w-48 text-xs sm:text-sm',
              props.settings.translation_enabled &&
              props.settings.translation_provider === 'baidu' &&
              !props.settings.baidu_app_id?.trim()
                ? 'border-red-500'
                : '',
            ]"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  baidu_app_id: (e.target as HTMLInputElement).value,
                })
            "
          />
        </div>
        <div class="sub-setting-item">
          <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhKey :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-sm">
                {{ t('baiduSecretKey') }} <span class="text-red-500">*</span>
              </div>
              <div class="text-xs text-text-secondary hidden sm:block">
                {{ t('baiduSecretKeyDesc') }}
              </div>
            </div>
          </div>
          <input
            :value="props.settings.baidu_secret_key"
            type="password"
            :placeholder="t('baiduSecretKeyPlaceholder')"
            :class="[
              'input-field w-32 sm:w-48 text-xs sm:text-sm',
              props.settings.translation_enabled &&
              props.settings.translation_provider === 'baidu' &&
              !props.settings.baidu_secret_key?.trim()
                ? 'border-red-500'
                : '',
            ]"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  baidu_secret_key: (e.target as HTMLInputElement).value,
                })
            "
          />
        </div>
      </template>

      <!-- AI Translation Settings -->
      <template v-if="props.settings.translation_provider === 'ai'">
        <div class="sub-setting-item">
          <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhKey :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-sm">
                {{ t('aiApiKey') }} <span class="text-red-500">*</span>
              </div>
              <div class="text-xs text-text-secondary hidden sm:block">
                {{ t('aiApiKeyDesc') }}
              </div>
            </div>
          </div>
          <input
            :value="props.settings.ai_api_key"
            type="password"
            :placeholder="t('aiApiKeyPlaceholder')"
            :class="[
              'input-field w-32 sm:w-48 text-xs sm:text-sm',
              props.settings.translation_enabled &&
              props.settings.translation_provider === 'ai' &&
              !props.settings.ai_api_key?.trim()
                ? 'border-red-500'
                : '',
            ]"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  ai_api_key: (e.target as HTMLInputElement).value,
                })
            "
          />
        </div>
        <div class="sub-setting-item">
          <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhLink :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('aiEndpoint') }}</div>
              <div class="text-xs text-text-secondary hidden sm:block">
                {{ t('aiEndpointDesc') }}
              </div>
            </div>
          </div>
          <input
            :value="props.settings.ai_endpoint"
            type="text"
            :placeholder="t('aiEndpointPlaceholder')"
            class="input-field w-32 sm:w-48 text-xs sm:text-sm"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  ai_endpoint: (e.target as HTMLInputElement).value,
                })
            "
          />
        </div>
        <div class="sub-setting-item">
          <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhRobot :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('aiModel') }}</div>
              <div class="text-xs text-text-secondary hidden sm:block">
                {{ t('aiModelDesc') }}
              </div>
            </div>
          </div>
          <input
            :value="props.settings.ai_model"
            type="text"
            :placeholder="t('aiModelPlaceholder')"
            class="input-field w-32 sm:w-48 text-xs sm:text-sm"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  ai_model: (e.target as HTMLInputElement).value,
                })
            "
          />
        </div>
        <div class="sub-setting-item flex-col items-stretch gap-2">
          <div class="flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhRobot :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('aiSystemPrompt') }}</div>
              <div class="text-xs text-text-secondary hidden sm:block">
                {{ t('aiSystemPromptDesc') }}
              </div>
            </div>
          </div>
          <textarea
            :value="props.settings.ai_system_prompt"
            class="input-field w-full text-xs sm:text-sm resize-none"
            rows="3"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  ai_system_prompt: (e.target as HTMLTextAreaElement).value,
                })
            "
          />
        </div>

        <!-- AI Usage Display -->
        <div class="sub-setting-item flex-col items-stretch gap-2">
          <div class="flex items-center justify-between gap-2 sm:gap-3 min-w-0">
            <div class="flex items-center gap-2 sm:gap-3">
              <PhChartLine :size="20" class="text-text-secondary shrink-0 sm:w-6 sm:h-6" />
              <div class="flex-1 min-w-0">
                <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('aiUsage') }}</div>
                <div class="text-xs text-text-secondary hidden sm:block">
                  {{ t('aiUsageTokensDesc') }}
                </div>
              </div>
            </div>
            <button
              type="button"
              class="flex items-center gap-1 px-2 py-1 text-xs rounded bg-bg-tertiary hover:bg-bg-secondary border border-border transition-colors"
              @click="resetAIUsage"
            >
              <PhArrowCounterClockwise :size="14" />
              {{ t('aiUsageReset') }}
            </button>
          </div>
          <div class="flex flex-col gap-2 mt-2">
            <!-- Usage bar -->
            <div class="flex items-center gap-2">
              <span class="text-xs text-text-secondary w-20">{{ t('aiUsageTokens') }}:</span>
              <span class="text-sm font-medium">
                {{ aiUsage.usage.toLocaleString() }}
                <span v-if="aiUsage.limit > 0" class="text-text-secondary">
                  / {{ aiUsage.limit.toLocaleString() }}
                </span>
                <span v-else class="text-text-secondary text-xs">({{ t('unlimited') }})</span>
              </span>
            </div>
            <!-- Progress bar (only shown if limit is set) -->
            <div
              v-if="aiUsage.limit > 0"
              class="relative h-2 bg-bg-tertiary rounded-full overflow-hidden"
            >
              <div
                class="absolute top-0 left-0 h-full transition-all duration-300 rounded-full"
                :class="aiUsage.limit_reached ? 'bg-red-500' : 'bg-accent'"
                :style="{ width: getUsagePercentage() + '%' }"
              />
            </div>
            <!-- Limit reached warning -->
            <div v-if="aiUsage.limit_reached" class="text-xs text-red-500 flex items-center gap-1">
              <span>⚠️</span>
              {{ t('aiUsageLimitReached') }}
            </div>
          </div>
        </div>

        <!-- AI Usage Limit Setting -->
        <div class="sub-setting-item">
          <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhChartLine :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('aiUsageLimit') }}</div>
              <div class="text-xs text-text-secondary hidden sm:block">
                {{ t('aiUsageLimitDesc') }}
              </div>
            </div>
          </div>
          <input
            :value="props.settings.ai_usage_limit"
            type="number"
            min="0"
            :placeholder="t('aiUsageLimitPlaceholder')"
            class="input-field w-32 sm:w-48 text-xs sm:text-sm"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  ai_usage_limit: (e.target as HTMLInputElement).value,
                })
            "
          />
        </div>
      </template>

      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhGlobe :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('targetLanguage') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('targetLanguageDesc') || 'Language to translate article titles to' }}
            </div>
          </div>
        </div>
        <select
          :value="props.settings.target_language"
          class="input-field w-24 sm:w-48 text-xs sm:text-sm"
          @change="
            (e) =>
              emit('update:settings', {
                ...props.settings,
                target_language: (e.target as HTMLSelectElement).value,
              })
          "
        >
          <option value="en">{{ t('english') }}</option>
          <option value="es">{{ t('spanish') }}</option>
          <option value="fr">{{ t('french') }}</option>
          <option value="de">{{ t('german') }}</option>
          <option value="zh">{{ t('chinese') }}</option>
          <option value="ja">{{ t('japanese') }}</option>
        </select>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors;
}
.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}
.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}
.toggle:checked::after {
  transform: translateX(20px);
}
.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}
.sub-setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-2.5 rounded-md bg-bg-tertiary;
}
</style>
