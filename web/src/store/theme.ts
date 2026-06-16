import { create } from 'zustand';
import { themeConfigs, defaultTheme } from '../themes';
import type { ThemeName, ThemeConfig } from '../themes/types';

const STORAGE_KEY = 'reviewbuddy-theme';

function applyTheme(name: ThemeName) {
  document.documentElement.setAttribute('data-theme', name);
}

interface ThemeStore {
  currentTheme: ThemeName;
  themeConfig: ThemeConfig;
  setTheme: (name: ThemeName) => void;
}

const initial = (localStorage.getItem(STORAGE_KEY) as ThemeName) || defaultTheme.name;
applyTheme(initial);

export const useThemeStore = create<ThemeStore>((set) => ({
  currentTheme: initial,
  themeConfig: themeConfigs[initial],
  setTheme: (name) => {
    applyTheme(name);
    localStorage.setItem(STORAGE_KEY, name);
    set({ currentTheme: name, themeConfig: themeConfigs[name] });
  },
}));
