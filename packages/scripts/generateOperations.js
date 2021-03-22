const fs = require('fs');
const path = require('path');

const {gqlPluckFromCodeStringSync} = require('@graphql-tools/graphql-tag-pluck');
const {sync: globSync} = require('glob');

const args = process.argv.slice(2);

if (args.length === 0) {
  console.error('You must provide an output file!');
}

const outputFile = args[0];

const operations = globSync(
  path.join(process.cwd(), '/src/**/+(queries|mutations|fragments)/*.ts'),
).reduce((acc, filePath) => {
  acc +=
    gqlPluckFromCodeStringSync(filePath, fs.readFileSync(filePath, 'utf8')) + '\n';

  return acc;
}, '');

fs.writeFileSync(
  outputFile,
  operations,
);
