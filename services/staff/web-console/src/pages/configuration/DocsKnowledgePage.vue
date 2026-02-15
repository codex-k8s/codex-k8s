<template>
  <div>
    <PageHeader :title="t('pages.docs.title')" :hint="t('pages.docs.hint')" />

    <VRow class="mt-4" density="compact">
      <VCol cols="12" md="3">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.docs.sidebarTitle") }}</VCardTitle>
          <VCardText class="pt-0">
            <VList density="compact">
              <VListItem v-for="n in nodes" :key="n.id" :title="n.title" prepend-icon="mdi-file-document-outline" />
            </VList>
          </VCardText>
        </VCard>
      </VCol>

      <VCol cols="12" md="6">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.docs.contentTitle") }}</VCardTitle>
          <VCardText>
            <VAlert type="info" variant="tonal" class="mb-3">
              <div class="text-body-2">{{ t("pages.docs.scaffoldNote") }}</div>
            </VAlert>

            <MonacoEditor v-model="markdown" language="markdown" height="520px" />
          </VCardText>
        </VCard>
      </VCol>

      <VCol cols="12" md="3">
        <VCard variant="outlined">
          <VCardTitle class="text-subtitle-2">{{ t("pages.docs.tocTitle") }}</VCardTitle>
          <VCardText class="pt-0">
            <VList density="compact">
              <VListItem v-for="h in toc" :key="h" :title="h" prepend-icon="mdi-format-list-bulleted" />
            </VList>
          </VCardText>
        </VCard>
      </VCol>
    </VRow>
  </div>
</template>

<script setup lang="ts">
// TODO(#19): Реализовать docs/knowledge: дерево + контент + TOC, code-blocks с copy, и markdown editor на Monaco Editor.
import { ref } from "vue";
import { useI18n } from "vue-i18n";

import PageHeader from "../../shared/ui/PageHeader.vue";
import MonacoEditor from "../../shared/ui/monaco/MonacoEditor.vue";

const { t } = useI18n({ useScope: "global" });

const nodes = [
  { id: "overview", title: "Overview" },
  { id: "labels", title: "Labels & stages" },
  { id: "agents", title: "Agents operating model" },
  { id: "mcp", title: "MCP approvals" },
];

const toc = ["# Title", "## Section", "## Another section"];

const markdown = ref("# Hello\n\nMonaco editor is enabled for markdown.");
</script>
