module.exports = {
  module: {
    rules: [
      {
        loader: require.resolve('file-loader'),
        include: [/\.(svg|webp)$/],
        options: {
          outputPath(url, resourcePath) {
            const temp = resourcePath.split('/');
            return `${temp[temp.length - 2]}/${temp[temp.length - 1]}`;
          }
        }
      }
    ]
  }
};
