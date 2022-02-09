import path from 'path';

import svgr from '@svgr/rollup';
import {defineConfig} from 'vite';
import tsconfigPaths from 'vite-tsconfig-paths';

import packageJson from './package.json';

const isExternal = (id: string) =>
  Object.keys(packageJson.peerDependencies).includes(id);

export default defineConfig(() => ({
  build: {
    lib: {
      entry: path.resolve(__dirname, 'src/index.ts'),
      formats: ['es', 'cjs'],
    },
    rollupOptions: {
      external: isExternal,
      output: {
        // Since we publish our ./src folder, there's no point
        // in bloating sourcemaps with another copy of it.
        sourcemapExcludeSources: true,
      },
    },
    sourcemap: true,
    target: 'ES2018',
    // Leave minification up to applications.
    minify: false,
  },
  plugins: [tsconfigPaths(), svgr()],
}));
