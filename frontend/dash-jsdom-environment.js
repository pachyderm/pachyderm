'use strict';

const DefaultJsdomEnvironment = require('jest-environment-jsdom');

class DashJsdomEnvironment extends DefaultJsdomEnvironment {
  constructor(config, options) {
    super({
      ...config,
      globals: {
        ...config.globals,
        Uint8Array: Uint8Array,
      },
      options,
    });
  }
}

module.exports = DashJsdomEnvironment;
