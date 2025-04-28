<template>
  <component :is="type" v-bind="linkProps">
    <slot />
  </component>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { isExternal } from "@/utils/validate";

const props = defineProps({
  to: {
    type: String,
    required: true,
  },
});

const isExternalLink = computed(() => {
  return isExternal(props.to);
});

const type = computed(() => {
  if (isExternalLink.value) {
    return "a";
  }
  return "router-link";
});

const linkProps = computed(() => {
  if (isExternalLink.value) {
    return {
      href: props.to,
      target: "_blank",
      rel: "noopener",
    };
  }
  return {
    to: props.to,
  };
});
</script>
