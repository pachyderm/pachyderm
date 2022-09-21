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
    lib: {
      entry: path.resolve(__dirname, 'src/index.ts'),
      name: "Components",
      fileName: 'components'
    },
    rollupOptions: {
      external: ['react', 'react-dom', 'react-router-dom'],
      output: {
        // Since we publish our ./src folder, there's no point
        // in bloating sourcemaps with another copy of it.
        sourcemapExcludeSources: true,
        globals: {
          react: 'React',
          ['react-dom']: 'ReactDOM',
          ['react-router-dom']: 'reactRouterDom'
        }
      },
    },
    sourcemap: true,
    target: 'ES2018',
    // Leave minification up to applications.
    minify: false,
  },
  css: {
    postcss: {
      plugins: [customMedia(), normalize(), flexbugFixes(), presetEnv()],
    },
  },
  plugins: [tsconfigPaths(), svgr()],
}));
