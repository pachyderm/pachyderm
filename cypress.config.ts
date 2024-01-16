import {defineConfig} from 'cypress';
import fs from 'fs';
import path from 'path';
import codeCoverageTask from '@cypress/code-coverage/task';

// function fromDir(startPath, filter) {
//   if (!fs.existsSync(startPath)) {
//     console.log('no dir ', startPath);
//     return null;
//   }

//   var files = fs.readdirSync(startPath);
//   for (var i = 0; i < files.length; i++) {
//     var filename = path.join(startPath, files[i]);
//     var stat = fs.lstatSync(filename);
//     if (stat.isDirectory()) {
//       const out = fromDir(filename, filter); //recurse
//       if (out) return out;
//     } else if (filter.test(filename)) return filename;
//   }

//   return null;
// }

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
        readDownloadedFileMaybe: (filename) => {
          const downloadsFolder = config.downloadsFolder;
          const filepath = path.join(downloadsFolder, filename);
          if (fs.existsSync(filepath)) {
            return fs.readFileSync(filepath, 'utf8');
          }
          return null;
        },
        deleteDownloadedFile: (filename) => {
          const downloadsFolder = config.downloadsFolder;
          const filepath = path.join(downloadsFolder, filename);
          if (fs.existsSync(filepath)) {
            fs.unlinkSync(filepath);
          }
          return true;
        },
        // deleteDownloadedFiles: () => {
        //   const fs = require('fs');
        //   const path = require('path');

        //   const directory = config.downloadsFolder;

        //   if (!fs.existsSync(directory)) {
        //     return true;
        //   }

        //   fs.readdir(directory, (err, files) => {
        //     if (err) throw err;

        //     for (const file of files) {
        //       fs.unlink(path.join(directory, file), (err) => {
        //         if (err) throw err;
        //       });
        //     }
        //   });

        //   return true;
        // },
        // checkZipFiles: () => {
        // // npm install node-stream-zip
        //   const StreamZip = require('node-stream-zip');

        //   const downloadsFolder = config.downloadsFolder;
        //   const filepath = fromDir(downloadsFolder, /pachyderm-download-/i);

        //   if (!filepath) {
        //     return null;
        //   }

        //   const zip = new StreamZip({
        //     file: filepath,
        //     storeEntries: true,
        //   });

        //   const files: any[] = [];

        //   return new Promise((res) => {
        //     zip.on('ready', () => {
        //       for (const entry of Object.values(zip.entries())) {
        //         files.push((entry as any).name);
        //       }

        //       zip.close();
        //       return res(files as string[]);
        //     });
        //   });
        // },
      });

      return config;
    },
  },
});
