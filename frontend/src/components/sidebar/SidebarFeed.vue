<script setup lang="ts">
import { ref } from 'vue';
import { PhWarningCircle, PhEyeSlash, PhImage, PhDotsSixVertical } from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

interface Props {
  feed: Feed;
  isActive: boolean;
  unreadCount: number;
  isDragging?: boolean;
  isEditMode?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  click: [];
  contextmenu: [event: MouseEvent];
  dragstart: [event: Event];
  dragend: [];
}>();

const isDraggingSelf = ref(false);

function getFavicon(url: string): string {
  try {
    return `https://www.google.com/s2/favicons?domain=${new URL(url).hostname}`;
  } catch {
    return '';
  }
}

function handleDragStart(event: Event) {
  isDraggingSelf.value = true;
  emit('dragstart', event);
}

function handleDragEnd() {
  isDraggingSelf.value = false;
  emit('dragend');
}
</script>

<template>
  <div
    :class="['feed-item', isActive ? 'active' : '', isDraggingSelf ? 'dragging' : '']"
    :data-feed-id="feed.id"
    @click="emit('click')"
    @contextmenu="(e) => emit('contextmenu', e)"
  >
    <!-- Drag handle (only visible in edit mode) -->
    <div
      v-if="isEditMode"
      class="drag-handle"
      draggable="true"
      :title="t('dragToReorder')"
      @dragstart="handleDragStart"
      @dragend="handleDragEnd"
    >
      <PhDotsSixVertical :size="14" />
    </div>

    <div class="w-4 h-4 flex items-center justify-center shrink-0">
      <img
        :src="feed.image_url || getFavicon(feed.url)"
        class="w-full h-full object-contain"
        @error="($event.target as HTMLElement).style.display = 'none'"
      />
    </div>
    <span class="truncate flex-1">{{ feed.title }}</span>
    <PhImage
      v-if="feed.is_image_mode"
      :size="16"
      class="text-accent shrink-0"
      :title="t('imageMode')"
    />
    <PhEyeSlash
      v-if="feed.hide_from_timeline"
      :size="16"
      class="text-text-secondary shrink-0"
      :title="t('hideFromTimeline')"
    />
    <PhWarningCircle
      v-if="feed.last_error"
      :size="16"
      class="text-yellow-500 shrink-0"
      :title="feed.last_error"
    />
    <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount }}</span>
  </div>
</template>

<style scoped>
@reference "../../style.css";

.feed-item {
  @apply px-2 sm:px-3 py-1.5 sm:py-2 cursor-pointer rounded-md text-xs sm:text-sm text-text-primary flex items-center gap-1.5 sm:gap-2.5 hover:bg-bg-tertiary transition-colors;
}
.feed-item.active {
  @apply bg-bg-tertiary text-accent font-medium;
}
.feed-item.dragging {
  opacity: 0.5;
  transform: scale(0.98);
}

.drag-handle {
  @apply cursor-grab text-text-secondary hover:text-accent transition-colors flex items-center justify-center;
  padding: 2px;
  margin-right: 2px;
  border-radius: 2px;
}
.drag-handle:hover {
  background-color: var(--color-bg-tertiary, rgba(0, 0, 0, 0.05));
}
.drag-handle:active {
  cursor: grabbing;
}

.unread-badge {
  @apply text-[9px] sm:text-[10px] font-semibold rounded-full min-w-[14px] sm:min-w-[16px] h-[14px] sm:h-[16px] px-0.5 sm:px-1 flex items-center justify-center;
  background-color: rgba(120, 120, 120, 0.25);
  color: #444444;
}
</style>

<style>
.dark-mode .unread-badge {
  background-color: rgba(100, 100, 100, 0.6) !important;
  color: #f0f0f0 !important;
}
</style>
