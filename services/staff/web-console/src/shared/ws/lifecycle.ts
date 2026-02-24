type RealtimePageLifecycleHandlers = {
  onResume: () => void;
  onSuspend?: () => void;
};

export function bindRealtimePageLifecycle(handlers: RealtimePageLifecycleHandlers): () => void {
  if (typeof window === "undefined" || typeof document === "undefined") {
    return () => undefined;
  }

  const handleResume = (): void => {
    if (document.hidden) return;
    handlers.onResume();
  };

  const handleVisibilityChange = (): void => {
    if (document.hidden) {
      handlers.onSuspend?.();
      return;
    }
    handlers.onResume();
  };

  window.addEventListener("focus", handleResume);
  window.addEventListener("online", handleResume);
  window.addEventListener("pageshow", handleResume);
  document.addEventListener("visibilitychange", handleVisibilityChange);

  return () => {
    window.removeEventListener("focus", handleResume);
    window.removeEventListener("online", handleResume);
    window.removeEventListener("pageshow", handleResume);
    document.removeEventListener("visibilitychange", handleVisibilityChange);
  };
}
