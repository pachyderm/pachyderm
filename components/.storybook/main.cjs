const {mergeConfig, loadConfigFromFile} = require('vite');
const path = require('path');

module.exports = {
  async viteFinal(config) {
    const { config: userConfig } = await loadConfigFromFile(
      path.resolve(__dirname, "../vite.config.ts")
    );
    userConfig.build.lib = undefined;
    userConfig.build.rollupOptions = {
      external: [],
      output: {},
    };
    return mergeConfig(config, userConfig);
  },
  stories: ['../src/**/*.stories.tsx'],
  addons: [
    "@storybook/addon-essentials",
    '@storybook/addon-links',
  ],
  framework: "@storybook/react",
  core: {
    builder: "@storybook/builder-vite"
  },
  staticDirs: ['../public'],
  features: {
    "storyStoreV7": true
  },
};
