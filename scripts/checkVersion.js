/* eslint-disable @typescript-eslint/no-var-requires */

// The last part of the npm version is only for npm usage
const packageVersion = require('../backend/node_modules/@pachyderm/node-pachyderm/package.json').dependencies[
  '@pachyderm/proto'
].replace(/\.\d+$/, '');
const versionVersion = require('../version.json').pachyderm;

if (packageVersion !== versionVersion) {
  console.error(
    `version.json(${versionVersion}) does not match package.json(${packageVersion})`,
  );
  process.exit(1);
}
