import {Version} from '@dash-backend/proto/proto/version/versionpb/version_pb';

const admin = new Version()
  .setMajor(2)
  .setMinor(5)
  .setMicro(4)
  .setPlatform('arm64')
  .setGoVersion('g01.19')
  .setGitTreeModified('false')
  .setGitCommit('143fff8a572a2ec019dce29363f9030b3599e1ff')
  .setBuildDate('2023-04-13T16:13:48Z')
  .setAdditional('additional');

export default admin;
