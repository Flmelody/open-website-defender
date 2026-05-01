export const THEME_STORAGE_KEY = "owd.theme";

const themeIds = new Set(["matrix", "cyan", "amber", "violet", "crimson"]);

const normalizeTheme = (theme: string | null) => {
  return theme && themeIds.has(theme) ? theme : "matrix";
};

export const initTheme = () => {
  const theme = normalizeTheme(localStorage.getItem(THEME_STORAGE_KEY));
  document.documentElement.dataset.theme = theme;
};

export const getThemeColor = (name: string, fallback: string) => {
  return (
    getComputedStyle(document.documentElement).getPropertyValue(name).trim() ||
    fallback
  );
};
