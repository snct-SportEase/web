

/** @type { import('@storybook/sveltekit').StorybookConfig } */
const config = {
  "stories": [
    "../src/**/*.stories.@(js|ts|svelte)"
  ],
  "addons": [
    "@storybook/addon-svelte-csf",
    "@storybook/addon-a11y",
    "@storybook/addon-docs"
  ],
  "framework": "@storybook/sveltekit"
};
export default config;
