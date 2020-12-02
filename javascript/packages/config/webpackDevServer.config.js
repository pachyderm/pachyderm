/* eslint-disable @typescript-eslint/no-var-requires */

const fs = require('fs');
const path = require('path');

const errorOverlayMiddleware = require('react-dev-utils/errorOverlayMiddleware');
const evalSourceMapMiddleware = require('react-dev-utils/evalSourceMapMiddleware');
const ignoredFiles = require('react-dev-utils/ignoredFiles');
const noopServiceWorkerMiddleware = require('react-dev-utils/noopServiceWorkerMiddleware');
const redirectServedPath = require('react-dev-utils/redirectServedPathMiddleware');

const host = process.env.HOST || '0.0.0.0';
const sockHost = process.env.WDS_SOCKET_HOST;
const sockPath = process.env.WDS_SOCKET_PATH; // default: '/sockjs-node'
const sockPort = process.env.WDS_SOCKET_PORT;
const appDir = fs.realpathSync(process.cwd());

module.exports = (proxy, allowedHost) => {
  return {
    disableHostCheck:
      !proxy || process.env.DANGEROUSLY_DISABLE_HOST_CHECK === 'true',
    compress: true,
    clientLogLevel: 'none',
    contentBase: path.resolve(appDir, 'public'),
    contentBasePublicPath: '/',
    watchContentBase: true,
    hot: true,
    transportMode: 'ws',
    injectClient: false,
    sockHost,
    sockPath,
    sockPort,
    publicPath: '',
    quiet: true,
    watchOptions: {
      ignored: ignoredFiles(path.resolve(appDir, 'src')),
    },
    host,
    overlay: false,
    historyApiFallback: {
      disableDotRule: true,
      index: '/',
    },
    public: allowedHost,
    proxy,
    before(app, server) {
      app.use(evalSourceMapMiddleware(server));
      app.use(errorOverlayMiddleware());
    },
    after(app) {
      app.use(redirectServedPath('/'));
      app.use(noopServiceWorkerMiddleware('/'));
    },
  };
};
