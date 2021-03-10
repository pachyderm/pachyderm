/* eslint-disable @typescript-eslint/no-var-requires,@typescript-eslint/camelcase */

const fs = require('fs');
const path = require('path');

const ReactRefreshWebpackPlugin = require('@pmmmwh/react-refresh-webpack-plugin');
const CaseSensitivePathsPlugin = require('case-sensitive-paths-webpack-plugin');
const CopyPlugin = require('copy-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const OptimizeCSSAssetsPlugin = require('optimize-css-assets-webpack-plugin');
const postcssNormalize = require('postcss-normalize');
const postcssSafeParser = require('postcss-safe-parser');
const ForkTsCheckerWebpackPlugin = require('react-dev-utils/ForkTsCheckerWebpackPlugin');
const getCSSModuleLocalIdent = require('react-dev-utils/getCSSModuleLocalIdent');
const InlineChunkHtmlPlugin = require('react-dev-utils/InlineChunkHtmlPlugin');
const InterpolateHtmlPlugin = require('react-dev-utils/InterpolateHtmlPlugin');
const ModuleNotFoundPlugin = require('react-dev-utils/ModuleNotFoundPlugin');
const typescriptFormatter = require('react-dev-utils/typescriptFormatter');
const WatchMissingNodeModulesPlugin = require('react-dev-utils/WatchMissingNodeModulesPlugin');
const TerserPlugin = require('terser-webpack-plugin');
const webpack = require('webpack');
const {BundleAnalyzerPlugin} = require('webpack-bundle-analyzer');

const appDir = fs.realpathSync(process.cwd());
const NODE_ENV = process.env.NODE_ENV;
const analyzeBuild = process.env.ANALYZE_WEBPACK_BUILD;
const isProd = NODE_ENV === 'production' || NODE_ENV === 'staging';
const isDev = NODE_ENV === 'development' || NODE_ENV === 'test';
const env = Object.keys(process.env)
  .filter((key) => key.startsWith('REACT_APP') || key === 'NODE_ENV')
  .reduce(
    (env, key) => {
      env.raw[key] = process.env[key];
      env.stringified[key] = JSON.stringify(process.env[key]);

      return env;
    },
    {raw: {}, stringified: {}},
  );

const postcssPlugins = () => [
  require('postcss-flexbugs-fixes'),
  require('postcss-preset-env')({
    autoprefixer: {
      flexbox: 'no-2009',
    },
    stage: 3,
  }),
  require('postcss-custom-media'),
  postcssNormalize(),
];

module.exports = {
  mode: isProd ? 'production' : 'development',
  bail: isProd,
  devtool: isProd ? 'source-map' : 'cheap-module-source-map',
  entry: [
    isDev && require.resolve('react-dev-utils/webpackHotDevClient'),
    path.resolve(appDir, 'src'),
  ].filter(Boolean),
  output: {
    path: isProd ? path.resolve(appDir, 'build') : undefined,
    pathinfo: isDev,
    filename: isDev
      ? 'static/js/bundle.js'
      : 'static/js/[name].[contenthash:8].js',
    futureEmitAssets: true,
    chunkFilename: isDev
      ? 'static/js/[name].chunk.js'
      : 'static/js/[name].[contenthash:8].chunk.js',
    publicPath: '/',
    jsonpFunction: `webpackJsonpHub`,
    globalObject: 'this',
  },
  performance: {
    hints: false,
  },
  optimization: {
    minimize: isProd,
    minimizer: [
      new TerserPlugin({
        terserOptions: {
          parse: {
            ecma: 8,
          },
          compress: {
            ecma: 5,
            warnings: false,
            comparisons: false,
            inline: 2,
          },
          mangle: {
            safari10: true,
          },
          keep_classnames: false,
          keep_fnames: false,
          output: {
            ascii_only: true,
            comments: false,
            ecma: 5,
          },
        },
        sourceMap: isProd,
      }),
      new OptimizeCSSAssetsPlugin({
        parser: postcssSafeParser,
        map: {
          annotation: true,
          inline: false,
        },
        cssProcessorPluginOptions: {
          preset: 'default',
        },
      }),
    ],
    splitChunks: {
      chunks: 'all',
      name: false,
    },
    runtimeChunk: {
      name: (entrypoint) => `runtime-${entrypoint.name}`,
    },
  },
  resolve: {
    extensions: ['.mjs', '.ts', '.tsx', '.js', '.jsx', '.json'],
    modules: ['node_modules', path.resolve(appDir, 'src')],
  },
  module: {
    rules: [
      {
        oneOf: [
          {
            test: [/\.bmp$/, /\.gif$/, /\.jpe?g$/, /\.png$/],
            loader: require.resolve('url-loader'),
            options: {
              limit: 10000,
              name: 'static/media/[name].[hash:8].[ext]',
            },
          },
          {
            test: /\.(js|mjs|jsx|ts|tsx)$/,
            include: [path.resolve(appDir, 'src')],
            loader: require.resolve('babel-loader'),
            options: {
              plugins: [
                [
                  require.resolve('babel-plugin-named-asset-import'),
                  {
                    loaderMap: {
                      svg: {
                        ReactComponent:
                          '@svgr/webpack?-svgo,+titleProp,+ref![path]',
                      },
                    },
                  },
                ],
                isDev && [require.resolve('react-refresh/babel')],
                isDev && [require.resolve('babel-plugin-istanbul')],
              ].filter(Boolean),
              cacheDirectory: true,
              cacheCompression: false,
              compact: isProd,
            },
          },
          {
            test: /\.css$/,
            exclude: /\.module\.css$/,
            use: [
              isProd
                ? {
                    loader: MiniCssExtractPlugin.loader,
                  }
                : require.resolve('style-loader'),
              {
                loader: require.resolve('css-loader'),
                options: {
                  importLoaders: 1,
                  sourceMap: isProd,
                },
              },
              {
                loader: require.resolve('postcss-loader'),
                options: {
                  ident: 'postcss',
                  plugins: postcssPlugins,
                  sourceMap: isProd,
                },
              },
            ],
            sideEffects: true,
          },
          {
            test: /\.module\.css$/,
            use: [
              isProd
                ? {
                    loader: MiniCssExtractPlugin.loader,
                  }
                : require.resolve('style-loader'),
              {
                loader: require.resolve('css-loader'),
                options: {
                  importLoaders: 1,
                  modules: {
                    getLocalIdent: getCSSModuleLocalIdent,
                  },
                  sourceMap: isProd,
                },
              },
              {
                loader: require.resolve('postcss-loader'),
                options: {
                  ident: 'postcss',
                  plugins: postcssPlugins,
                  sourceMap: isProd,
                },
              },
            ],
          },
          {
            loader: require.resolve('file-loader'),
            exclude: [/\.(js|mjs|jsx|ts|tsx)$/, /\.html$/, /\.json$/],
            options: {
              name: 'static/media/[name].[hash:8].[ext]',
            },
          },
          // Add all new loaders before the file-loader
        ],
      },
    ],
    strictExportPresence: true,
  },
  plugins: [
    new HtmlWebpackPlugin(
      Object.assign(
        {},
        {
          inject: true,
          template: path.resolve(appDir, 'public/index.html'),
        },
        isProd
          ? {
              minify: {
                removeComments: true,
                collapseWhitespace: true,
                removeRedundantAttributes: true,
                useShortDoctype: true,
                removeEmptyAttributes: true,
                removeStyleLinkTypeAttributes: true,
                keepClosingSlash: true,
                minifyJS: true,
                minifyCSS: true,
                minifyURLs: true,
              },
            }
          : undefined,
      ),
    ),
    new CopyPlugin({
      patterns: [
        {
          from: path.resolve(appDir, 'public'),
          to: path.resolve(appDir, 'build'),
          globOptions: {
            ignore: ['index.html'],
          },
        },
      ],
    }),
    isProd && new InlineChunkHtmlPlugin(HtmlWebpackPlugin, [/runtime-.+[.]js/]),
    new InterpolateHtmlPlugin(HtmlWebpackPlugin, env.raw),
    new webpack.DefinePlugin({'process.env': env.stringified}),
    new ModuleNotFoundPlugin(appDir),
    isDev && new webpack.HotModuleReplacementPlugin(),
    isDev && new ReactRefreshWebpackPlugin(),
    isDev && new CaseSensitivePathsPlugin(),
    isDev &&
      new WatchMissingNodeModulesPlugin(path.resolve(appDir, 'node_modules')),
    isProd &&
      new MiniCssExtractPlugin({
        filename: 'static/css/[name].[contenthash:8].css',
        chunkFilename: 'static/css/[name].[contenthash:8].chunk.css',
      }),
    new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/),
    new ForkTsCheckerWebpackPlugin({
      typescript: require.resolve('typescript'),
      async: isDev,
      useTypescriptIncrementalApi: true,
      checkSyntacticErrors: true,
      tsconfig: path.resolve(appDir, 'tsconfig.json'),
      reportFiles: [
        '**',
        '!**/__tests__/**',
        '!**/?(*.)(spec|test).*',
        '!**/src/setupProxy.*',
        '!**/src/setupTests.*',
      ],
      formatter: isProd ? typescriptFormatter : undefined,
    }),
    analyzeBuild && new BundleAnalyzerPlugin(),
  ].filter(Boolean),
  node: {
    module: 'empty',
    dgram: 'empty',
    dns: 'mock',
    fs: 'empty',
    http2: 'empty',
    net: 'empty',
    tls: 'empty',
    child_process: 'empty',
  },
};
