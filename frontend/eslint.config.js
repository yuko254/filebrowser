import js from "@eslint/js";
import pluginVue from "eslint-plugin-vue";
import tseslintPlugin from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import prettierConfig from "eslint-config-prettier";
import prettierPlugin from "eslint-plugin-prettier";

export default [
  // Global ignores
  {
    ignores: ["dist/**", "node_modules/**", "*.config.js", ".*lintrc.js"],
  },

  // Base ESLint recommended rules
  js.configs.recommended,

  // TypeScript configuration
  {
    files: ["**/*.{ts,tsx,vue}"],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: "latest",
        sourceType: "module",
        extraFileExtensions: [".vue"],
      },
    },
    plugins: {
      "@typescript-eslint": tseslintPlugin,
    },
    rules: {
      // Include recommended TypeScript rules manually
      ...tseslintPlugin.configs.recommended.rules,
      // You can override rules here if needed
    },
  },

  // Vue configuration (uses the Vue plugin's recommended flat config)
  ...pluginVue.configs["flat/recommended"],

  // Ensure Vue files use TypeScript parser
  {
    files: ["**/*.vue"],
    languageOptions: {
      parserOptions: {
        parser: tsParser,
      },
    },
  },

  // Additional custom rules (adjust as needed)
  {
    rules: {
      "vue/multi-word-component-names": "warn",
      "vue/no-mutating-props": "warn",
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/no-unused-expressions": "warn",
      "no-undef": "warn",
      "no-unsafe-finally": "warn",
      "no-useless-assignment": "warn",
      "no-empty": "warn",
    },
  },
];