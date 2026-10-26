<script>
export default {
  name: "SettingsHeader",
  props: {
    title: { type: String, required: true },
    description: { type: String, default: "" },
    note: { type: String, default: "" },
  },
};
</script>

<template>
  <header class="settings-heading">
    <span class="settings-heading-icon" aria-hidden="true"><slot /></span>
    <div class="settings-heading-copy">
      <h2>{{ title }}</h2>
      <p v-if="description">{{ description }}</p>
    </div>
    <span v-if="note" class="settings-heading-note">{{ note }}</span>
  </header>
</template>
