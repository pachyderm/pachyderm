/* eslint-disable @typescript-eslint/no-var-requires */
const fs = require('fs');
const path = require('path');
const {Readable} = require('stream');

const uniq = require('lodash/uniq');

const {
  modifyFetchStreamingErrorHandling,
} = require('./modify-fetch-streaming-error-handling');
const {addTypeNamesToTypes} = require('./strengthen-types');

const OUT_DIR = './src/generated/proto';
const GOOGLE_TYPES_DIR = './@types/google/protobuf';
const PACH_VERSION = 'master';
const PACH_PATH = `https://raw.githubusercontent.com/pachyderm/pachyderm/${PACH_VERSION}/src/typescript`;
const PACHYDERM_MODULES = [
  'admin/admin',
  'auth/auth',
  'debug/debug',
  'enterprise/enterprise',
  'identity/identity',
  'license/license',
  'pfs/pfs',
  'pps/pps',
  'protoextensions/log',
  'protoextensions/validate',
  'protoextensions/json-schema-options',
  'proxy/proxy',
  'task/task',
  'transaction/transaction',
  'version/versionpb/version',
  'worker/worker',
];
const FOLDERS = uniq([
  ...PACHYDERM_MODULES.map((m) => {
    const parts = m.split('/');

    return parts.length < 3 ? parts[0] : parts.slice(0, 2).join('/');
  }),
  'google/protobuf',
]);

const downloadFile = async (filePath) => {
  const fullFilePath = `${filePath}.pb.ts`;
  const response = await fetch(`${PACH_PATH}/${fullFilePath}`);
  const fileStream = fs.createWriteStream(path.resolve(OUT_DIR, fullFilePath));
  const pipe = Readable.fromWeb(response.body).pipe(fileStream);

  return new Promise((resolve, reject) => {
    pipe.on('finish', resolve);
    pipe.on('error', reject);
  });
};

const downloadTypes = async () => {
  // Clean up old folders
  fs.rmSync(OUT_DIR, {recursive: true, force: true});

  // Make folders for the types
  FOLDERS.forEach((f) => {
    fs.mkdirSync(path.resolve(OUT_DIR, f), {recursive: true});
  });

  // Download pachyderm types
  const downloads = PACHYDERM_MODULES.map(downloadFile);

  // Download fetch helper
  await downloadFile('fetch');

  // Copy over google/protobuf types
  fs.readdirSync(path.resolve(GOOGLE_TYPES_DIR)).forEach((filePath) => {
    fs.copyFileSync(
      path.resolve(GOOGLE_TYPES_DIR, filePath),
      path.resolve(OUT_DIR, 'google/protobuf', filePath),
    );
  });

  // Wait for all the proto files to be downloaded
  await Promise.all(downloads);

  modifyFetchStreamingErrorHandling();
  // Strengthen all types
  addTypeNamesToTypes();
};

downloadTypes();
