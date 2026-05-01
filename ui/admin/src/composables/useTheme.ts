import { ref } from "vue";

export type ThemeId = "matrix" | "cyan" | "amber" | "violet" | "crimson";

export interface ThemeOption {
  id: ThemeId;
  label: string;
  accent: string;
}

export const THEME_STORAGE_KEY = "owd.theme";

export const themeOptions: ThemeOption[] = [
  { id: "matrix", label: "MATRIX", accent: "#00ff00" },
  { id: "cyan", label: "CYAN", accent: "#00e5ff" },
  { id: "amber", label: "AMBER", accent: "#ffb000" },
  { id: "violet", label: "VIOLET", accent: "#b87cff" },
  { id: "crimson", label: "CRIMSON", accent: "#ff4d6d" },
];

const defaultTheme: ThemeId = "matrix";
const themeIds = new Set(themeOptions.map((theme) => theme.id));

const normalizeTheme = (theme: string | null): ThemeId => {
  return theme && themeIds.has(theme as ThemeId)
    ? (theme as ThemeId)
    : defaultTheme;
};

const readStoredTheme = (): ThemeId => {
  if (typeof window === "undefined") return defaultTheme;
  return normalizeTheme(window.localStorage.getItem(THEME_STORAGE_KEY));
};

const currentTheme = ref<ThemeId>(readStoredTheme());

export const applyTheme = (theme: string, persist = true) => {
  const nextTheme = normalizeTheme(theme);
  currentTheme.value = nextTheme;

  if (typeof document !== "undefined") {
    document.documentElement.dataset.theme = nextTheme;
  }

  if (typeof window !== "undefined") {
    if (persist) {
      window.localStorage.setItem(THEME_STORAGE_KEY, nextTheme);
    }
    window.dispatchEvent(
      new CustomEvent("owd-theme-change", { detail: nextTheme }),
    );
  }
};

export const initTheme = () => {
  applyTheme(readStoredTheme(), false);
};

export const useTheme = () => ({
  themeOptions,
  currentTheme,
  setTheme: applyTheme,
});

const cssVar = (name: string, fallback: string) => {
  if (typeof document === "undefined") return fallback;
  const value = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  return value || fallback;
};

export const getThemeColors = () => {
  const accentRgb = cssVar("--theme-accent-rgb", "0, 255, 0");
  const panelRgb = cssVar("--theme-panel-rgb", "10, 25, 10");

  return {
    accent: cssVar("--theme-accent", "#00ff00"),
    accentStrong: cssVar("--theme-accent-strong", "#00cc00"),
    accentRgb,
    border: cssVar("--theme-accent-border", "#005000"),
    borderSoft: cssVar("--theme-border-soft", "#004000"),
    borderFaint: cssVar("--theme-border-faint", "#002800"),
    text: cssVar("--theme-text", "#ccffcc"),
    textDim: cssVar("--theme-text-dim", "#88aa88"),
    panel: cssVar("--theme-bg-panel", "rgba(10, 25, 10, 0.85)"),
    tooltipBg: `rgba(${panelRgb}, 0.94)`,
    areaStart: `rgba(${accentRgb}, 0.3)`,
    areaEnd: `rgba(${accentRgb}, 0.02)`,
    subtleFill: `rgba(${accentRgb}, 0.15)`,
  };
};
