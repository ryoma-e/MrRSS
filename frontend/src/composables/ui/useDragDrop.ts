import { ref, type Ref } from 'vue';
import type { Feed } from '@/types/models';

export interface DragState {
  draggingFeedId: number | null;
  dragOverCategory: string | null;
  dragOverFeedId: number | null;
  dropPosition: 'before' | 'after' | 'inside' | null;
}

export function useDragDrop() {
  const draggingFeedId: Ref<number | null> = ref(null);
  const dragOverCategory: Ref<string | null> = ref(null);
  const dragOverFeedId: Ref<number | null> = ref(null);
  const dropPosition: Ref<'before' | 'after' | 'inside' | null> = ref(null);

  function onDragStart(feedId: number, event: Event) {
    const dragEvent = event as DragEvent;
    draggingFeedId.value = feedId;
    if (dragEvent.dataTransfer) {
      dragEvent.dataTransfer.effectAllowed = 'move';
      dragEvent.dataTransfer.setData('text/plain', String(feedId));
    }
  }

  function onDragEnd() {
    draggingFeedId.value = null;
    dragOverCategory.value = null;
    dragOverFeedId.value = null;
    dropPosition.value = null;
  }

  function onDragOver(category: string, feedId: number | null, event: Event) {
    const dragEvent = event as DragEvent;
    dragEvent.preventDefault();
    if (!draggingFeedId.value) return;

    // Don't allow dropping on itself
    if (feedId === draggingFeedId.value) return;

    dragOverCategory.value = category;
    dragOverFeedId.value = feedId;

    // Calculate drop position based on mouse Y position relative to element
    if (feedId !== null && dragEvent.target instanceof HTMLElement) {
      const target = dragEvent.currentTarget as HTMLElement;
      const rect = target.getBoundingClientRect();
      const relativeY = dragEvent.clientY - rect.top;
      const threshold = rect.height / 2;

      if (relativeY < threshold) {
        dropPosition.value = 'before';
      } else {
        dropPosition.value = 'after';
      }
    } else {
      dropPosition.value = 'inside';
    }
  }

  function onDragLeave() {
    // Only clear if we're actually leaving (not just hovering over a child)
    // This is handled by checking if we're still over a valid target
  }

  async function onDrop(
    currentCategory: string,
    feeds: Feed[]
  ): Promise<{ success: boolean; error?: string }> {
    if (!draggingFeedId.value) {
      return { success: false, error: 'No feed being dragged' };
    }

    const feedId = draggingFeedId.value;
    const targetCategory = dragOverCategory.value || currentCategory;
    const targetFeedId = dragOverFeedId.value;

    // Find the target feed in the list to determine position
    let newPosition = 0;
    if (targetFeedId !== null) {
      const targetFeed = feeds.find((f) => f.id === targetFeedId);
      if (targetFeed) {
        // If dropping before, use the target's position
        // If dropping after, use target's position + 1
        newPosition = targetFeed.position || 0;
        if (dropPosition.value === 'after') {
          newPosition += 1;
        }
      }
    } else {
      // No specific feed target, append to end of category
      const maxPos = Math.max(0, ...feeds.map((f) => f.position || 0));
      newPosition = maxPos + 1;
    }

    try {
      const response = await fetch('/api/feeds/reorder', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          feed_id: feedId,
          category: targetCategory,
          position: newPosition,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to reorder feed');
      }

      return { success: true };
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
    }
  }

  function getDropIndicatorClass(feedId: number): string {
    if (dragOverFeedId.value !== feedId) return '';
    return dropPosition.value === 'before'
      ? 'drop-before'
      : dropPosition.value === 'after'
        ? 'drop-after'
        : '';
  }

  function getCategoryDropClass(category: string): string {
    if (dragOverCategory.value === category && dropPosition.value === 'inside') {
      return 'category-drop-target';
    }
    return '';
  }

  return {
    draggingFeedId,
    dragOverCategory,
    dragOverFeedId,
    dropPosition,
    onDragStart,
    onDragEnd,
    onDragOver,
    onDragLeave,
    onDrop,
    getDropIndicatorClass,
    getCategoryDropClass,
  };
}
