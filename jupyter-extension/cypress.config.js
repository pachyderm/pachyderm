const { defineConfig } = require('cypress')

module.exports = defineConfig({
  e2e: {
    baseUrl: "http://localhost:8888/",
    defaultCommandTimeout: 20000,
    projectId: "dttaoj",
  },
  video: true,
});
