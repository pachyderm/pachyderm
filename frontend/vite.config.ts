import path from 'path';

import react from '@vitejs/plugin-react';
import analyze from 'rollup-plugin-analyzer';
import copy from 'rollup-plugin-copy';
import {defineConfig} from 'vite';
import {ViteEjsPlugin} from 'vite-plugin-ejs';
import EnvironmentPlugin from 'vite-plugin-environment';
import svgr from 'vite-plugin-svgr';
import tsconfigPaths from 'vite-tsconfig-paths';

const NODE_ENV = process.env.NODE_ENV;

const isProd = NODE_ENV === 'production' || NODE_ENV === 'staging';

const env = Object.keys(process.env)
  .filter((key) => key.startsWith('REACT_APP') || key === 'NODE_ENV')
  .reduce(
    (env: Record<string, any>, key) => {
      if (key.startsWith('REACT_APP_RUNTIME')) {
        env.pachDashConfig[key] = isProd
          ? `window.pachDashConfig['${key}']`
          : process.env[key] || '';
      } else {
        env[key] = process.env[key] || '';
      }

      return env;
    },
    {pachDashConfig: {}},
  );

// https://vitejs.dev/config/
export default defineConfig({
  // vite will automatically set this to "development" if you are using the dev server.
  mode: process.env.NODE_ENV,
  server: {
    proxy: {
      '/graphql': 'http://localhost:3000',
      '/upload': 'http://localhost:3000',
    },
    port: 4000,
  },
  // This changes the out put dir from dist to build
  // comment this out if that isn't relevant for your project
  build: {
    outDir: 'build',
  },
  resolve: {
    alias: {
      '@pachyderm/components': path.resolve(__dirname, '/components/src'),
      '@pachyderm/components/*': path.resolve(__dirname, '/components/src/*'),
    },
  },
  plugins: [
    react(),
    svgr(),
    tsconfigPaths(),
    EnvironmentPlugin({...env}),
    ViteEjsPlugin(),
    copy({
      targets: [{src: 'oauth/**/*', dest: 'build/oauth'}],
      hook: 'writeBundle',
    }),
    ...(process.env.ANALYZE ? [analyze()] : []),
  ],
});
