const glob = require('glob');
const path = require('path');

module.exports = (env) => {
  return {
    mode: 'production',
    entry: glob.sync('./tests/*.ts').reduce((acc, match) => {
      const entry = path.basename(match, '.ts');
      if (!env.REST || (!!env.REST && entry === 'rest')) {
        acc[entry] = './' + match;
      }
      return acc;
    }, {}),
    output: {
      path: path.resolve(__dirname, 'build'),
      libraryTarget: 'commonjs',
      filename: '[name].bundle.js',
    },
    resolve: {
      extensions: ['.ts', '.js'],
      alias: {
        '@queries': path.resolve(__dirname, '../frontend/src/queries'),
        '@mutations': path.resolve(__dirname, '../frontend/src/mutations'),
        '@dash-frontend': path.resolve(__dirname, '../frontend/src/'),
      },
    },
    module: {
      rules: [
        {
          test: /\.ts$/,
          exclude: !!env.REST
            ? [
                path.resolve(__dirname, './tests/graphql.ts'),
                path.resolve(__dirname, './utils.ts'),
              ]
            : undefined,
          use: 'babel-loader',
        },
      ],
    },
    target: 'web',
    externals: /k6(\/.*)?/,
  };
};
