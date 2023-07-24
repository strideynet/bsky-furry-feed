module.exports = {
  env: {
    browser: true,
    es2021: true,
    node: true,
  },
  extends: ["eslint:recommended", "@nuxt/eslint-config"],
  rules: {
    "vue/multi-word-component-names": "off",
    "vue/html-self-closing": "off",
    "vue/max-attributes-per-line": "off",
    "vue/singleline-html-element-content-newline": "off",
    "vue/html-indent": "off",
    "vue/html-closing-bracket-newline": "off",
  },
};
