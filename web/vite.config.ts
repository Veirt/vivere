import { fileURLToPath, URL } from "node:url";

import { presets } from "@effect/tsgo/oxlint-presets";
import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import { devtools } from "@tanstack/devtools-vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import { redact } from "@tanstack/redact/vite";
import viteReact, { reactCompilerPreset } from "@vitejs/plugin-react";
import { Record } from "effect";
import { nitro } from "nitro/vite";
import { defineConfig } from "vite-plus";

const ignorePatterns = [
  ".nitro/**",
  ".output/**",
  ".tanstack/**",
  "dist/**",
  "public/**",
  "src/routeTree.gen.ts",
];

const effectTsgoRules = Record.map(
  {
    ...presets.antipattern.rules,
    ...presets.correctness.rules,
    ...presets.effectNative.rules,
    ...presets.recommended.rules,
    ...presets.style.rules,
    ...presets.strict.rules,
  },
  () => "error" as const,
);

const isVitest = process.env.VITEST !== undefined;

export default defineConfig({
  ...(isVitest ? {} : { ssr: { noExternal: true } }),
  resolve: {
    tsconfigPaths: true,
    ...(isVitest
      ? {
          conditions: ["module"],
          mainFields: ["module", "jsnext:main", "jsnext", "main"],
        }
      : {}),
    alias: {
      "#": fileURLToPath(new URL("./src", import.meta.url)),
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  plugins: [
    ...(isVitest
      ? []
      : [
          devtools(),
          nitro({
            rollupConfig: { external: [/^@sentry\//] },
          }),
          tanstackStart(),
        ]),
    tailwindcss(),
    viteReact(),
    babel({ presets: [reactCompilerPreset()] }),
    redact(),
  ],
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    include: ["src/**/*.test.{ts,tsx}"],
    server: {
      deps: {
        inline: [
          /@testing-library\/react/,
          /radix-ui/,
          /@radix-ui\//,
          /react-remove-scroll/,
          /react-style-singleton/,
          /use-callback-ref/,
          /use-sidecar/,
          /@floating-ui\/react-dom/,
          /lucide-react/,
          /@tanstack\/react-/,
          /use-sync-external-store/,
        ],
      },
    },
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
        files: ["vite.config.ts"],
        rules: {
          "effecttsgo/process-env": "off",
        },
      },
      {
        files: ["src/routes/**/*.tsx", "src/components/**/*.tsx", "src/**/*.test.{ts,tsx}"],
        rules: {
          "effecttsgo/strict-boolean-expressions": "off",
          "effecttsgo/prefer-effect-array": "off",
          "effecttsgo/prefer-effect-string": "off",
          "effecttsgo/prefer-effect-record": "off",
        },
      },
      {
        files: ["src/routes/**/*.tsx", "src/components/ui/**/*.tsx"],
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
