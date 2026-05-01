<template>
  <div class="initial-table-loader no-select" aria-hidden="true">
    <div class="loader-spinner" />
  </div>
</template>

<style scoped>
.initial-table-loader {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 64px 16px 72px;
  border-top: 1px solid var(--theme-border-faint);
  background:
    linear-gradient(180deg, rgba(var(--theme-panel-rgb), 0.18), rgba(var(--theme-panel-rgb), 0.08)),
    repeating-linear-gradient(
      0deg,
      rgba(var(--theme-accent-rgb), 0.03),
      rgba(var(--theme-accent-rgb), 0.03) 1px,
      transparent 1px,
      transparent 24px
    );
  color: var(--theme-accent);
  font-family: "Courier New", monospace;
  letter-spacing: 0.04em;
}

.loader-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(var(--theme-accent-rgb), 0.2);
  border-top-color: var(--theme-accent);
  border-radius: 50%;
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
