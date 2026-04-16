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
  border-top: 1px solid #003000;
  background:
    linear-gradient(180deg, rgba(0, 40, 0, 0.18), rgba(0, 20, 0, 0.08)),
    repeating-linear-gradient(
      0deg,
      rgba(0, 255, 0, 0.03),
      rgba(0, 255, 0, 0.03) 1px,
      transparent 1px,
      transparent 24px
    );
  color: #0f0;
  font-family: "Courier New", monospace;
  letter-spacing: 0.04em;
}

.loader-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(0, 255, 0, 0.2);
  border-top-color: #0f0;
  border-radius: 50%;
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
