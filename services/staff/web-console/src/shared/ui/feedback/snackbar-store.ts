import { defineStore } from "pinia";

export type SnackbarKind = "success" | "error" | "info";

export type SnackbarItem = {
  id: number;
  text: string;
  kind: SnackbarKind;
  timeoutMs: number;
};

let nextId = 1;

export const useSnackbarStore = defineStore("snackbar", {
  state: () => ({
    queue: [] as SnackbarItem[],
    current: null as SnackbarItem | null,
    open: false,
  }),
  actions: {
    push(text: string, kind: SnackbarKind = "info", timeoutMs = 3500): void {
      this.queue.push({
        id: nextId++,
        text,
        kind,
        timeoutMs,
      });
      if (!this.open) {
        this.openNext();
      }
    },
    success(text: string, timeoutMs = 2500): void {
      this.push(text, "success", timeoutMs);
    },
    error(text: string, timeoutMs = 4500): void {
      this.push(text, "error", timeoutMs);
    },
    info(text: string, timeoutMs = 3500): void {
      this.push(text, "info", timeoutMs);
    },
    openNext(): void {
      const next = this.queue.shift() || null;
      this.current = next;
      this.open = Boolean(next);
    },
    close(): void {
      this.open = false;
      this.current = null;
      this.openNext();
    },
  },
});

