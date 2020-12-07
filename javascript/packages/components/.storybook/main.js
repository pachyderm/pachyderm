const path = require('path');

const baseConfig = require('../webpack.config.js');

module.exports = {
  stories: ['../src/**/*.stories.tsx'],
  webpackFinal: (config) => {
    const cssModuleRule = baseConfig.module.rules[0].oneOf.find(
      (rule) => rule.test.toString() === '/\\.module\\.css$/'
    );
    const cssRule = config.module.rules.find(
      (rule) => rule.test.toString() === '/\\.css$/'
    );

    cssRule.exclude = cssModuleRule.test;
    config.module.rules.push(cssModuleRule);
    config.plugins.push(...baseConfig.plugins);

    return config;
  },
  addons: [
    '@storybook/addon-actions',
    "@storybook/addon-essentials",
    '@storybook/addon-links'
  ],
};
