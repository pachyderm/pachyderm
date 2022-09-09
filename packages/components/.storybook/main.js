const {mergeConfig, loadConfigFromFile} = require('vite');
const path = require('path');

module.exports = {
  async viteFinal(config) {
    const { config: userConfig } = await loadConfigFromFile(
      path.resolve(__dirname, "../vite.config.ts")
    );
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
