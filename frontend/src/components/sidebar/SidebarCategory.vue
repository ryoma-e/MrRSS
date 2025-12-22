<script setup lang="ts">
import { PhFolder, PhFolderDashed, PhCaretDown } from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';
import type { DropPreview } from '@/composables/ui/useDragDrop';
import SidebarFeed from './SidebarFeed.vue';

interface Props {
  name: string;
  feeds: Feed[];
  isOpen: boolean;
  isActive: boolean;
  isUncategorized?: boolean;
  unreadCount: number;
  currentFeedId: number | null;
  feedUnreadCounts: Record<number, number>;
  isDragOver?: boolean;
  isEditMode?: boolean;
  dropPreview?: DropPreview;
}

defineProps<Props>();

const emit = defineEmits<{
  toggle: [];
  selectCategory: [];
  selectFeed: [feedId: number];
  categoryContextMenu: [event: MouseEvent];
  feedContextMenu: [event: MouseEvent, feed: Feed];
  feedDragOver: [feedId: number | null, event: Event];
  drop: [];
  dragstart: [feedId: number, event: Event];
  dragend: [];
}>();

// Handle dragover on the feeds-list container using event delegation
function handleFeedsListDragOver(event: DragEvent) {
  // Prevent default to allow drop
  event.preventDefault();

  // Find which feed item we're hovering over
  const target = event.target as HTMLElement;
  const feedItem = target.closest('.feed-item');

  if (feedItem) {
    // Get the feed ID from the data attribute
    const feedIdStr = feedItem.getAttribute('data-feed-id');
    const feedId = feedIdStr ? parseInt(feedIdStr, 10) : null;
    console.log('[SidebarCategory] Emitting feedDragOver with feedId:', feedId, 'event:', event);
    emit('feedDragOver', feedId, event);
  } else {
    // Not hovering over any specific feed
    console.log('[SidebarCategory] Emitting feedDragOver with null feedId, event:', event);
    emit('feedDragOver', null, event);
  }
}

function handleDrop() {
  emit('drop');
}

// Handle dragover on category container (for dropping at category level)
function handleCategoryDragOver(event: DragEvent) {
  event.preventDefault();
  emit('feedDragOver', null, event);
}
</script>

<template>
  <div
    :class="['mb-1 category-container', isDragOver ? 'drag-over' : '']"
    @dragover.self="handleCategoryDragOver"
    @dragleave.self="() => {}"
    @drop.self="handleDrop"
  >
    <div
      :class="['category-header', isActive ? 'active' : '']"
      @contextmenu="(e) => emit('categoryContextMenu', e)"
    >
      <span class="flex-1 flex items-center gap-2" @click="emit('selectCategory')">
        <PhFolderDashed v-if="isUncategorized" :size="20" />
        <PhFolder v-else :size="20" :weight="'fill'" />
        {{ name }}
      </span>
      <span v-if="unreadCount > 0" class="unread-badge mr-1">{{ unreadCount }}</span>
      <PhCaretDown
        :size="20"
        class="p-1 cursor-pointer transition-transform"
        :class="{ 'rotate-180': isOpen }"
        @click.stop="emit('toggle')"
      />
    </div>
    <div
      v-show="isOpen"
      class="pl-2 feeds-list"
      @dragover="handleFeedsListDragOver"
      @drop.prevent="handleDrop"
    >
      <template v-for="feed in feeds" :key="feed.id">
        <!-- Drop indicator above this feed -->
        <div
          v-if="
            isDragOver &&
            dropPreview &&
            dropPreview.targetFeedId === feed.id &&
            dropPreview.beforeTarget
          "
          class="drop-indicator"
        ></div>
        <SidebarFeed
          :feed="feed"
          :is-active="currentFeedId === feed.id"
          :unread-count="feedUnreadCounts[feed.id] || 0"
          :is-edit-mode="isEditMode"
          @click="emit('selectFeed', feed.id)"
          @contextmenu="(e) => emit('feedContextMenu', e, feed)"
          @dragstart="(e) => emit('dragstart', feed.id, e)"
          @dragend="emit('dragend')"
        />
        <!-- Drop indicator below this feed -->
        <div
          v-if="
            isDragOver &&
            dropPreview &&
            dropPreview.targetFeedId === feed.id &&
            !dropPreview.beforeTarget
          "
          class="drop-indicator"
        ></div>
      </template>
      <!-- Drop indicator at the end when dragging over category but not over a specific feed -->
      <div
        v-if="isDragOver && feeds.length > 0 && dropPreview && dropPreview.targetFeedId === null"
        class="drop-indicator"
      ></div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";

.category-header {
  @apply px-2 sm:px-3 py-1.5 sm:py-2 cursor-pointer font-semibold text-xs sm:text-sm text-text-secondary flex items-center justify-between rounded-md hover:bg-bg-tertiary hover:text-text-primary transition-colors;
  @apply sticky z-10 bg-bg-secondary;
  top: -0.375rem; /* matches container's p-1.5 */
  margin-left: -0.375rem;
  margin-right: -0.375rem;
  padding-left: calc(0.5rem + 0.375rem);
  padding-right: calc(0.75rem + 0.375rem);
}
@media (min-width: 640px) {
  .category-header {
    top: -0.5rem; /* matches container's sm:p-2 */
    margin-left: -0.5rem;
    margin-right: -0.5rem;
    padding-left: calc(0.75rem + 0.5rem);
    padding-right: calc(0.75rem + 0.5rem);
  }
}
.category-header.active {
  @apply bg-bg-tertiary text-accent;
}

.category-container.drag-over {
  @apply rounded-lg;
  border: 1px solid var(--color-accent, #6366f1);
  background-color: var(--color-bg-tertiary, rgba(99, 102, 241, 0.05));
}

.feeds-list {
  position: relative;
}

.drop-indicator {
  height: 3px;
  background: linear-gradient(90deg, transparent, var(--color-accent, #6366f1), transparent);
  margin: 2px 0;
  border-radius: 1.5px;
  animation: pulse-indicator 1.5s ease-in-out infinite;
}

@keyframes pulse-indicator {
  0%,
  100% {
    opacity: 0.6;
  }
  50% {
    opacity: 1;
  }
}

.unread-badge {
  @apply text-[9px] sm:text-[10px] font-semibold rounded-full min-w-[14px] sm:min-w-[16px] h-[14px] sm:h-[16px] px-0.5 sm:px-1 flex items-center justify-center;
  background-color: rgba(120, 120, 120, 0.25);
  color: #444444;
}
</style>

<style>
.dark-mode .unread-badge {
  /* This style will be applied to child components, so it can not use scoped */
  background-color: rgba(100, 100, 100, 0.6) !important;
  color: #f0f0f0 !important;
}
</style>
