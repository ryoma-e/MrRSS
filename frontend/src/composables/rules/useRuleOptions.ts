import { computed, type ComputedRef } from 'vue';
import { useAppStore } from '@/stores/app';

export interface Condition {
  id: number;
  logic?: 'and' | 'or' | null;
  negate: boolean;
  field: string;
  operator?: string | null;
  value: string;
  values: string[];
}

export interface FieldOption {
  value: string;
  labelKey: string;
  multiSelect: boolean;
  booleanField?: boolean;
}

export interface ActionOption {
  value: string;
  labelKey: string;
}

export function useRuleOptions() {
  const store = useAppStore();

  // Field options for conditions
  const fieldOptions: FieldOption[] = [
    { value: 'feed_name', labelKey: 'feedName', multiSelect: true },
    { value: 'feed_category', labelKey: 'feedCategory', multiSelect: true },
    { value: 'article_title', labelKey: 'articleTitle', multiSelect: false },
    { value: 'feed_type', labelKey: 'feedType', multiSelect: true },
    {
      value: 'is_freshrss_feed',
      labelKey: 'isFreshRSSFeed',
      multiSelect: false,
      booleanField: true,
    },
    {
      value: 'is_image_mode_feed',
      labelKey: 'isImageModeFeed',
      multiSelect: false,
      booleanField: true,
    },
    { value: 'published_after', labelKey: 'publishedAfter', multiSelect: false },
    { value: 'published_before', labelKey: 'publishedBefore', multiSelect: false },
    { value: 'is_read', labelKey: 'readStatus', multiSelect: false, booleanField: true },
    { value: 'is_favorite', labelKey: 'favoriteStatus', multiSelect: false, booleanField: true },
    { value: 'is_hidden', labelKey: 'hiddenStatus', multiSelect: false, booleanField: true },
    { value: 'is_read_later', labelKey: 'readLaterStatus', multiSelect: false, booleanField: true },
  ];

  // Operator options for article title
  const textOperatorOptions: Array<{ value: string; labelKey: string }> = [
    { value: 'contains', labelKey: 'contains' },
    { value: 'exact', labelKey: 'exactMatch' },
    { value: 'regex', labelKey: 'regex' },
  ];

  // Boolean value options
  const booleanOptions: Array<{ value: string; labelKey: string }> = [
    { value: 'true', labelKey: 'yes' },
    { value: 'false', labelKey: 'no' },
  ];

  // Action options
  const actionOptions: ActionOption[] = [
    { value: 'favorite', labelKey: 'actionFavorite' },
    { value: 'unfavorite', labelKey: 'actionUnfavorite' },
    { value: 'hide', labelKey: 'actionHide' },
    { value: 'unhide', labelKey: 'actionUnhide' },
    { value: 'mark_read', labelKey: 'actionMarkRead' },
    { value: 'mark_unread', labelKey: 'actionMarkUnread' },
    { value: 'read_later', labelKey: 'actionReadLater' },
    { value: 'remove_read_later', labelKey: 'actionRemoveReadLater' },
  ];

  // Feed names for multi-select
  const feedNames: ComputedRef<string[]> = computed(() => {
    return store.feeds.map((f) => f.title);
  });

  // Feed categories for multi-select
  const feedCategories: ComputedRef<string[]> = computed(() => {
    const categories = new Set<string>();
    store.feeds.forEach((f) => {
      if (f.category) {
        categories.add(f.category);
      }
    });
    return Array.from(categories);
  });

  // Feed types for multi-select
  const feedTypes: ComputedRef<string[]> = computed(() => {
    const types = new Set<string>();
    store.feeds.forEach((f) => {
      // Map frontend type to backend type
      if (f.type) {
        types.add(f.type);
      } else {
        // Empty type means regular RSS/Atom feed
        types.add('');
      }
    });
    return Array.from(types);
  });

  return {
    fieldOptions,
    textOperatorOptions,
    booleanOptions,
    actionOptions,
    feedNames,
    feedCategories,
    feedTypes,
  };
}

// Helper functions for field types
export function isDateField(field: string): boolean {
  return field === 'published_after' || field === 'published_before';
}

export function isMultiSelectField(field: string): boolean {
  return field === 'feed_name' || field === 'feed_category' || field === 'feed_type';
}

export function isBooleanField(field: string): boolean {
  return (
    field === 'is_read' ||
    field === 'is_favorite' ||
    field === 'is_hidden' ||
    field === 'is_read_later' ||
    field === 'is_freshrss_feed' ||
    field === 'is_image_mode_feed'
  );
}

export function needsOperator(field: string): boolean {
  return field === 'article_title';
}
