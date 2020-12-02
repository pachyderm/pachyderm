const env = process.env.BABEL_ENV || process.env.NODE_ENV;

module.exports = {
  presets: [
    [
      '@babel/preset-env',
      env === 'test'
        ? {
            targets: {
              node: 'current',
            },
          }
        : {
            corejs: 3,
            exclude: ['transform-typeof-symbol'],
            modules: false,
            useBuiltIns: 'entry',
          },
    ],
    [
      '@babel/preset-react',
      {
        development: env !== 'production',
        useBuiltIns: true,
      },
    ],
    '@babel/preset-typescript',
  ],
  plugins: [
    [
      '@babel/plugin-transform-runtime',
      {
        corej: false,
        helpers: true,
        regenerator: true,
        useESModules: false,
      },
    ],
    '@babel/plugin-proposal-optional-chaining',
    '@babel/plugin-proposal-nullish-coalescing-operator',
  ],
};
