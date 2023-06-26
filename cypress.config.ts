import {defineConfig} from 'cypress';
import fs from 'fs';
import path from 'path';
import codeCoverageTask from '@cypress/code-coverage/task';

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
      codeCoverageTask(on, config);
      config.env.AUTH_EMAIL = process.env.PACHYDERM_AUTH_EMAIL;
      config.env.AUTH_PASSWORD = process.env.PACHYDERM_AUTH_PASSWORD;

      on('task', {
        readFileMaybe: (filename) => {
          const downloadsFolder = config.downloadsFolder;
          const filepath = path.join(downloadsFolder, filename);
          if (fs.existsSync(filepath)) {
            return fs.readFileSync(filepath, 'utf8');
          }
          return null;
        },
      });

      return config;
    },
  },
});
