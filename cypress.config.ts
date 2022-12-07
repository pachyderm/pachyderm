import {defineConfig} from 'cypress';

export default defineConfig({
  chromeWebSecurity: false,
  defaultCommandTimeout: 6000,
  pageLoadTimeout: 120000,
  projectId: 'wameh2',
  screenshotsFolder: '/tmp/cypress-screenshots',
  trashAssetsBeforeRuns: true,
  videosFolder: '/tmp/cypress-videos',
  viewportHeight: 900, // default is 800
  viewportWidth: 1000, // default is 1000
  e2e: {
    baseUrl: 'http://localhost:4000/',
    specPattern: 'cypress/e2e/**/*.{js,jsx,ts,tsx}',
    setupNodeEvents(on, config) {
      return require('./cypress/plugins/index.js')(on, config);
    },
  },
  retries: {
    runMode: 2,
    openMode: 0,
  },
});
