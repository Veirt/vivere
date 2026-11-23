import { fileURLToPath, URL } from "node:url";

import { presets } from "@effect/tsgo/oxlint-presets";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { Record } from "effect";
import { defineConfig } from "vite-plus";

const ignorePatterns = [
  ".agent/**",
  ".agents/**",
  ".claude/**",
  ".codex/**",
  ".continue/**",
  ".cursor/**",
  ".gemini/**",
  ".opencode/**",
  ".pi/**",
  ".roo/**",
  ".windsurf/**",
  "dist/**",
  "public/**",
  "server/**",
  "src/routeTree.gen.ts",
];

const effectTsgoRules = Record.map(
  {
    ...presets.antipattern.rules,
    ...presets.correctness.rules,
    ...presets.effectNative.rules,
    ...presets.maintainers.rules,
    ...presets.recommended.rules,
    ...presets.style.rules,
    ...presets.strict.rules,
  },
  () => "error" as const,
);

export default defineConfig({
  plugins: [
    tanstackRouter({
      target: "react",
      autoCodeSplitting: true,
      routesDirectory: "./src/routes",
    }),
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:3000",
        changeOrigin: true,
      },
    },
  },
  test: {
    include: ["test/**/*.test.ts"],
  },
  lint: {
    ignorePatterns,
    env: {
      browser: true,
    },
    plugins: ["eslint", "typescript", "oxc", "react", "effecttsgo"],
    jsPlugins: [{ name: "vite-plus", specifier: "vite-plus/oxlint-plugin" }],
    options: {
      typeAware: true,
      typeCheck: true,
    },
    rules: {
      ...effectTsgoRules,
      "effecttsgo/missing-pipeable-signature": "off",
      "vite-plus/prefer-vite-plus-imports": "error",
      "oxc/no-accumulating-spread": "error",
      "eslint/no-debugger": "error",
      "eslint/no-duplicate-imports": "error",
      "eslint/no-unreachable": "error",
      "eslint/no-unused-vars": "error",
      "eslint/prefer-const": "error",
      "typescript/no-explicit-any": "error",
      "react/rules-of-hooks": "error",
      "react/exhaustive-deps": "error",
      "react/only-export-components": ["warn", { allowConstantExport: true }],
    },
    overrides: [
      {
        files: ["src/features/characters/api.ts"],
        rules: {
          "effecttsgo/global-fetch-in-effect": "off",
        },
      },
      {
        files: ["src/routes/**/*.tsx"],
        rules: {
          "react/only-export-components": "off",
        },
      },
    ],
  },
  fmt: {
    ignorePatterns,
    sortPackageJson: false,
  },
});
