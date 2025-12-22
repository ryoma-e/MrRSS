import { ref, type Ref } from 'vue';
import type { Feed } from '@/types/models';

export interface DropPreview {
  targetFeedId: number | null;
  beforeTarget: boolean; // true = insert before target, false = insert after target
}

export function useDragDrop() {
  const draggingFeedId: Ref<number | null> = ref(null);
  const dragOverCategory: Ref<string | null> = ref(null);
  const dropPreview: Ref<DropPreview> = ref({ targetFeedId: null, beforeTarget: true });

  function onDragStart(feedId: number, event: Event) {
    const dragEvent = event as DragEvent;
    draggingFeedId.value = feedId;
    if (dragEvent.dataTransfer) {
      dragEvent.dataTransfer.effectAllowed = 'move';
      dragEvent.dataTransfer.setData('text/plain', String(feedId));
    }
    console.log('[onDragStart] Started dragging feed:', feedId);
  }

  function onDragEnd() {
    console.log('[onDragEnd] Ended dragging feed:', draggingFeedId.value);
    draggingFeedId.value = null;
    dragOverCategory.value = null;
    dropPreview.value = { targetFeedId: null, beforeTarget: true };
  }

  function onDragOver(category: string, targetFeedId: number | null, event: Event) {
    if (!event || !(event instanceof DragEvent)) {
      console.log('[onDragOver] Invalid event:', event);
      return;
    }

    event.preventDefault();

    if (!draggingFeedId.value) {
      console.log('[onDragOver] No dragging feed');
      return;
    }

    // Don't allow dropping on itself
    if (targetFeedId === draggingFeedId.value) {
      dropPreview.value = { targetFeedId: null, beforeTarget: true };
      console.log('[onDragOver] Dropping on itself, clearing preview');
      return;
    }

    dragOverCategory.value = category;

    // Calculate drop position based on mouse Y position relative to element
    let beforeTarget = true;
    if (targetFeedId !== null && event.target instanceof HTMLElement) {
      // Use target instead of currentTarget to get the actual element being hovered
      const target = event.target;
      // Get the feed-item element (might need to traverse up)
      const feedItem = target.closest('.feed-item');
      if (feedItem) {
        const rect = feedItem.getBoundingClientRect();
        const relativeY = event.clientY - rect.top;
        const threshold = rect.height / 2;
        beforeTarget = relativeY < threshold;
        console.log(
          '[onDragOver] category:',
          category,
          'targetFeedId:',
          targetFeedId,
          'relativeY:',
          relativeY.toFixed(1),
          'threshold:',
          threshold.toFixed(1),
          'beforeTarget:',
          beforeTarget
        );
      } else {
        console.log('[onDragOver] Could not find .feed-item element');
      }
    } else {
      console.log('[onDragOver] No specific target, dropping at end. targetFeedId:', targetFeedId);
    }

    dropPreview.value = { targetFeedId, beforeTarget };
    console.log('[onDragOver] Updated dropPreview:', dropPreview.value);
  }

  function onDragLeave() {
    // Clear preview when leaving the category
    dropPreview.value = { targetFeedId: null, beforeTarget: true };
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
    const { targetFeedId, beforeTarget } = dropPreview.value;

    console.log('[onDrop] Starting drop operation:', {
      feedId,
      currentCategory,
      targetCategory,
      targetFeedId,
      beforeTarget,
      feedsCount: feeds.length,
    });

    // Sort feeds by position to get correct order
    const sortedFeeds = [...feeds].sort((a, b) => (a.position || 0) - (b.position || 0));

    // Calculate new position - simplified approach
    let newPosition = 0;

    if (targetFeedId !== null) {
      // Find target feed in the sorted list
      const targetIndex = sortedFeeds.findIndex((f) => f.id === targetFeedId);

      if (targetIndex !== -1) {
        const targetFeed = sortedFeeds[targetIndex];
        const targetPosition = targetFeed.position ?? targetIndex;

        // Simply use the target's position (before) or position+1 (after)
        // The backend ReorderFeed function will handle shifting other feeds correctly
        newPosition = beforeTarget ? targetPosition : targetPosition + 1;
      } else {
        // Target feed not found in this category's list (might be in different category)
        // Append to end
        const maxPos =
          sortedFeeds.length > 0 ? Math.max(0, ...sortedFeeds.map((f) => f.position || 0)) : -1;
        newPosition = maxPos + 1;
      }
    } else {
      // No specific feed target (dropping on empty space or category), append to end
      const maxPos =
        sortedFeeds.length > 0 ? Math.max(0, ...sortedFeeds.map((f) => f.position || 0)) : -1;
      newPosition = maxPos + 1;
    }

    console.log('[onDrop] Calculated position:', {
      feedId,
      targetCategory,
      newPosition,
      targetFeedId,
      beforeTarget,
    });

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
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to reorder feed');
      }

      console.log('[onDrop] Successfully reordered feed');
      return { success: true };
    } catch (error) {
      console.error('[onDrop] Error:', error);
      return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
    }
  }

  return {
    draggingFeedId,
    dragOverCategory,
    dropPreview,
    onDragStart,
    onDragEnd,
    onDragOver,
    onDragLeave,
    onDrop,
  };
}
