/* eslint-disable @typescript-eslint/no-var-requires */

// The last part of the npm version is only for npm usage
const packageVersion = require('../package.json').dependencies[
  '@pachyderm/proto'
].replace(/\.\d+$/, '');
const versionVersion = require('../version.json').pachyderm;

if (packageVersion !== versionVersion) {
  console.error(
    `version.json(${packageVersion}) does not match package.json(${versionVersion})`,
  );
  process.exit(1);
}
