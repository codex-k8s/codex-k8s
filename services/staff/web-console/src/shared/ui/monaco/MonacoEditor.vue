<template>
  <div class="root">
    <div ref="elRef" class="editor" :style="{ height }" />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";

type MonacoModule = typeof import("monaco-editor");

const props = withDefaults(
  defineProps<{
    modelValue: string;
    language: "markdown" | "yaml";
    height?: string;
    readOnly?: boolean;
  }>(),
  {
    height: "420px",
    readOnly: false,
  },
);

const emit = defineEmits<{
  (e: "update:modelValue", v: string): void;
}>();

const elRef = ref<HTMLElement | null>(null);

let monaco: MonacoModule | null = null;
let editor: import("monaco-editor").editor.IStandaloneCodeEditor | null = null;
let suppressEmit = false;

let workersReady = false;

async function ensureWorkers(): Promise<void> {
  if (workersReady) return;

  const [
    { default: EditorWorker },
    { default: JsonWorker },
    { default: CssWorker },
    { default: HtmlWorker },
    { default: TsWorker },
  ] = await Promise.all([
    import("monaco-editor/esm/vs/editor/editor.worker?worker"),
    import("monaco-editor/esm/vs/language/json/json.worker?worker"),
    import("monaco-editor/esm/vs/language/css/css.worker?worker"),
    import("monaco-editor/esm/vs/language/html/html.worker?worker"),
    import("monaco-editor/esm/vs/language/typescript/ts.worker?worker"),
  ]);

  // Vite: use getWorker (not getWorkerUrl)
  self.MonacoEnvironment = {
    getWorker: (_workerId: string, label: string) => {
      switch (label) {
        case "json":
          return new JsonWorker();
        case "css":
        case "scss":
        case "less":
          return new CssWorker();
        case "html":
        case "handlebars":
        case "razor":
          return new HtmlWorker();
        case "typescript":
        case "javascript":
          return new TsWorker();
        default:
          return new EditorWorker();
      }
    },
  } as unknown as typeof self.MonacoEnvironment;

  workersReady = true;
}

async function ensureMonaco(): Promise<MonacoModule> {
  if (monaco) return monaco;
  await ensureWorkers();

  // Load contributions only for needed languages.
  await Promise.all([
    import("monaco-editor/esm/vs/basic-languages/markdown/markdown.contribution"),
    import("monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution"),
  ]);

  monaco = await import("monaco-editor");
  return monaco;
}

async function mountEditor(): Promise<void> {
  const el = elRef.value;
  if (!el) return;

  const m = await ensureMonaco();
  editor = m.editor.create(el, {
    value: props.modelValue,
    language: props.language,
    readOnly: props.readOnly,
    automaticLayout: true,
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    wordWrap: "on",
  });

  editor.onDidChangeModelContent(() => {
    if (!editor) return;
    if (suppressEmit) return;
    emit("update:modelValue", editor.getValue());
  });
}

function setValue(next: string): void {
  if (!editor) return;
  const current = editor.getValue();
  if (current === next) return;

  suppressEmit = true;
  editor.setValue(next);
  suppressEmit = false;
}

function setLanguage(next: "markdown" | "yaml"): void {
  if (!monaco || !editor) return;
  const model = editor.getModel();
  if (!model) return;
  monaco.editor.setModelLanguage(model, next);
}

onMounted(() => void mountEditor());

watch(
  () => props.modelValue,
  (v) => setValue(v),
);

watch(
  () => props.language,
  (v) => setLanguage(v),
);

watch(
  () => props.readOnly,
  (v) => {
    if (!editor) return;
    editor.updateOptions({ readOnly: v });
  },
);

onBeforeUnmount(() => {
  editor?.dispose();
  editor = null;
});
</script>

<style scoped>
.root {
  border: 1px solid rgba(17, 24, 39, 0.12);
  border-radius: 12px;
  overflow: hidden;
  background: #fff;
}
.editor {
  width: 100%;
}
</style>

