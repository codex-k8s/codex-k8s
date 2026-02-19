import { ref, type Ref } from "vue";

const defaultPrefetchPages = 2;

type ProgressiveTableOptions = {
  itemsPerPage: number;
  prefetchPages?: number;
};

export type ProgressiveTableState = {
  page: Ref<number>;
  limit: Ref<number>;
  canLoadMore: Ref<boolean>;
  reset: () => void;
  markLoaded: (loadedCount: number) => void;
  shouldGrowForPage: (totalItems: number, nextPage: number, prevPage: number) => boolean;
};

export function createProgressiveTableState(options: ProgressiveTableOptions): ProgressiveTableState {
  const itemsPerPage = Math.max(1, Math.trunc(options.itemsPerPage));
  const prefetchPages = Math.max(1, Math.trunc(options.prefetchPages ?? defaultPrefetchPages));
  const step = itemsPerPage * prefetchPages;

  const page = ref(1);
  const limit = ref(step);
  const canLoadMore = ref(true);

  function reset(): void {
    page.value = 1;
    limit.value = step;
    canLoadMore.value = true;
  }

  function markLoaded(loadedCount: number): void {
    canLoadMore.value = loadedCount >= limit.value;
  }

  function shouldGrowForPage(totalItems: number, nextPage: number, prevPage: number): boolean {
    if (nextPage <= prevPage) {
      return false;
    }
    if (!canLoadMore.value) {
      return false;
    }
    if (totalItems <= 0) {
      return false;
    }
    const totalPages = Math.max(1, Math.ceil(totalItems / itemsPerPage));
    if (nextPage < totalPages) {
      return false;
    }
    limit.value += step;
    return true;
  }

  return {
    page,
    limit,
    canLoadMore,
    reset,
    markLoaded,
    shouldGrowForPage,
  };
}
