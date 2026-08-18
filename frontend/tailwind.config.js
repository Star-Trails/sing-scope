import daisyui from 'daisyui'

export default {
  darkMode: ['class', '[data-theme="dark"]'],
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'Helvetica Neue', 'Arial', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'Menlo', 'Monaco', 'Consolas', 'Liberation Mono', 'Courier New', 'monospace'],
      },
    },
  },
  plugins: [daisyui],
  daisyui: {
    themes: [
      {
        light: {
          "primary": "#6865c6",
          "primary-content": "#ffffff",
          "secondary": "#0071e3",
          "secondary-content": "#ffffff",
          "accent": "#34c759",
          "accent-content": "#ffffff",
          "neutral": "#3a3a3a",
          "neutral-content": "#f5f5f7",
          "base-100": "#ffffff",
          "base-200": "#f5f5f7",
          "base-300": "#e5e5ea",
          "base-content": "#1d1d1f",
          "info": "#5ac8fa",
          "success": "#34c759",
          "warning": "#ff9500",
          "error": "#ff3b30",
        },
        dark: {
          "primary": "#8583e9",
          "primary-content": "#ffffff",
          "secondary": "#2997ff",
          "secondary-content": "#ffffff",
          "accent": "#30d158",
          "accent-content": "#ffffff",
          "neutral": "#3a3a3c",
          "neutral-content": "#f5f5f7",
          "base-100": "#1c1c1e",
          "base-200": "#121214",
          "base-300": "#2c2c2e",
          "base-content": "#f5f5f7",
          "info": "#64d2ff",
          "success": "#30d158",
          "warning": "#ff9f0a",
          "error": "#ff453a",
        },
      },
      "synthwave",
      "dracula",
      "nord",
      "black",
    ],
    darkTheme: "dark",
    base: true,
    styled: true,
    utils: true,
  },
}
