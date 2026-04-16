import { computed, ref, watch, type Ref } from "vue";

export function useInitialTableLoading(loading: Ref<boolean>) {
  const hasLoadedOnce = ref(false);
  const sawFirstRequest = ref(loading.value);

  watch(loading, (isLoading) => {
    if (isLoading) {
      sawFirstRequest.value = true;
      return;
    }

    if (sawFirstRequest.value) {
      hasLoadedOnce.value = true;
    }
  });

  const initialLoading = computed(() => !hasLoadedOnce.value);

  return {
    hasLoadedOnce,
    initialLoading,
  };
}
