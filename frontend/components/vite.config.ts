import path from 'path';

import svgr from '@svgr/rollup';
import customMedia from 'postcss-custom-media';
import flexbugFixes from 'postcss-flexbugs-fixes';
import normalize from 'postcss-normalize';
import presetEnv from 'postcss-preset-env';
import {defineConfig} from 'vite';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig(() => ({
  build: {
    sourcemap: true,
    target: 'ES2018',
    // Leave minification up to applications.
    minify: false,
  },
  resolve: {
    alias: {
      '@pachyderm/components': path.resolve(__dirname, './src'),
      '@pachyderm/components/*': path.resolve(__dirname, '/components/src/*'),
    },
  },
  css: {
    postcss: {
      plugins: [customMedia(), normalize(), flexbugFixes(), presetEnv()],
    },
  },
  plugins: [tsconfigPaths(), svgr()],
}));
