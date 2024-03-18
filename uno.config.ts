// uno.config.ts
import { defineConfig, presetWind } from "unocss";

export default defineConfig({
  presets: [presetWind()],
  theme: {
    colors: {
      text: {
        50: "#f2f2f2",
        100: "#e6e6e6",
        200: "#cccccc",
        300: "#b3b3b3",
        400: "#999999",
        500: "#808080",
        600: "#666666",
        700: "#4d4d4d",
        800: "#333333",
        900: "#1a1a1a",
        950: "#0d0d0d",
      },
      background: {
        50: "#f2f2f2",
        100: "#e6e6e6",
        200: "#cccccc",
        300: "#b3b3b3",
        400: "#999999",
        500: "#808080",
        600: "#666666",
        700: "#4d4d4d",
        800: "#333333",
        900: "#1a1a1a",
        950: "#0d0d0d",
      },
      primary: {
        50: "#f6f3ef",
        100: "#ede7de",
        200: "#dbd0bd",
        300: "#c9b89c",
        400: "#b7a17b",
        500: "#a4895b",
        600: "#846e48",
        700: "#635236",
        800: "#423724",
        900: "#211b12",
        950: "#100e09",
      },
      secondary: {
        50: "#f2f2f3",
        100: "#e5e6e6",
        200: "#cbcccd",
        300: "#b1b3b4",
        400: "#97999b",
        500: "#7d8082",
        600: "#646668",
        700: "#4b4d4e",
        800: "#323334",
        900: "#191a1a",
        950: "#0c0d0d",
      },
      accent: {
        50: "#f2f3f2",
        100: "#e6e6e5",
        200: "#cdcdcb",
        300: "#b3b4b1",
        400: "#9a9b97",
        500: "#81827d",
        600: "#676864",
        700: "#4d4e4b",
        800: "#343432",
        900: "#1a1a19",
        950: "#0d0d0c",
      },
    },
  },
});
