/* eslint-disable @typescript-eslint/no-var-requires */

const path = require('path');

const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const postcssNormalize = require('postcss-normalize');
const getCSSModuleLocalIdent = require('react-dev-utils/getCSSModuleLocalIdent');
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');

const styleLoader =
  process.env.NODE_ENV === 'production'
    ? {
        loader: MiniCssExtractPlugin.loader,
      }
    : require.resolve('style-loader');

const postcssLoader = {
  loader: require.resolve('postcss-loader'),
  options: {
    ident: 'postcss',
    plugins: () => [
      require('postcss-flexbugs-fixes'),
      require('postcss-preset-env')({
        autoprefixer: {
          flexbox: 'no-2009',
        },
        stage: 3,
      }),
      postcssNormalize(),
    ],
    sourceMap: process.env.NODE_ENV === 'production',
  },
};

const config = {
  mode: 'production',
  devtool: 'hidden-source-map',
  entry: './src/index.ts',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: './index.js',
    libraryTarget: 'umd',
    globalObject: 'this',
  },
  module: {
    rules: [
      {
        oneOf: [
          {
            test: /\.(js|mjs|jsx|ts|tsx)$/,
            include: [path.resolve(__dirname, 'src')],
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
              ],
              cacheDirectory: true,
              cacheCompression: false,
              compact: true,
            },
          },
          {
            test: /\.css$/,
            exclude: /\.module\.css$/,
            use: [
              styleLoader,
              {
                loader: require.resolve('css-loader'),
                options: {
                  importLoaders: 1,
                  sourceMap: process.env.NODE_ENV === 'production',
                },
              },
              postcssLoader,
            ],
            sideEffects: true,
          },
          {
            test: /\.module\.css$/,
            use: [
              styleLoader,
              {
                loader: require.resolve('css-loader'),
                options: {
                  importLoaders: 1,
                  modules: {
                    getLocalIdent: getCSSModuleLocalIdent,
                  },
                  sourceMap: true,
                },
              },
              postcssLoader,
            ],
          },
          {
            loader: require.resolve('file-loader'),
            include: [/\.svg$/],
            options: {
              outputPath(url, resourcePath) {
                const temp = resourcePath.split('/');
                return `${temp[temp.length - 2]}/${temp[temp.length - 1]}`;
              },
            },
          },
        ],
      },
      // Add all new loaders before the file-loader
    ],
  },
  resolve: {
    extensions: ['.tsx', '.ts', '.jsx', '.js', '.json'],
    plugins: [new TsconfigPathsPlugin()],
  },
  externals: {
    react: 'react',
    'react-dom': 'react-dom',
    'react-router-dom': 'react-router-dom',
    'react-hook-form': 'react-hook-form',
    'react-helmet': 'react-helmet',
  },
  plugins: [
    new MiniCssExtractPlugin({
      filename: 'style.css',
      chunkFilename: 'style.[contenthash:8].chunk.css',
    }),
  ],
};

module.exports = config;
