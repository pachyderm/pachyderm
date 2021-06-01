import {FileInfo, FileType} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {fileInfoFromObject} from '@dash-backend/grpc/builders/pfs';

const tutorial = [
  fileInfoFromObject({
    committed: {seconds: 1614126189, nanos: 0},
    file: {
      commitId: 'd350c8d08a644ed5b2ee98c035ab6b33',
      path: '/AT-AT.png',
      branch: {name: 'master', repo: {name: 'images'}},
    },
    fileType: FileType.FILE,
    hash: 'P2fxZjakvux5dNsEfc0iCx1n3Kzo2QcDlzu9y3Ra1gc=',
    sizeBytes: 80588,
  }),
  fileInfoFromObject({
    committed: {seconds: 1610126189, nanos: 0},
    file: {
      commitId: 'd350c8d08a644ed5b2ee98c035ab6b33',
      path: '/liberty.png',
      branch: {name: 'master', repo: {name: 'images'}},
    },
    fileType: FileType.FILE,
    hash: 'QJoij3vijagMPZAadaQ1PtLRL7NFNZgouPoPQLbSi8E=',
    sizeBytes: 58644,
  }),
  fileInfoFromObject({
    committed: {seconds: 1612126189, nanos: 0},
    file: {
      commitId: 'd350c8d08a644ed5b2ee98c035ab6b33',
      path: '/kitten.png',
      branch: {name: 'master', repo: {name: 'images'}},
    },
    fileType: FileType.FILE,
    hash: 'AGyiZfGAxLyuqfB1yCGe/AMCpp2zdTYW8B37j8ls4hA',
    sizeBytes: 104836,
  }),
  fileInfoFromObject({
    committed: {seconds: 1616126189, nanos: 0},
    file: {
      commitId: 'd350c8d08a644ed5b2ee98c035ab6b33',
      path: '/cats/',
      branch: {name: 'master', repo: {name: 'images'}},
    },
    fileType: FileType.DIR,
    hash: 'sMAJnz5xBEFfYUVUo5PZrOpaoPZ902b+7N6+Fg5ACkQ=',
    sizeBytes: 98747,
  }),
];

const files: {[projectId: string]: FileInfo[]} = {
  '1': tutorial,
  '2': tutorial,
  '3': tutorial,
  '4': tutorial,
  '5': tutorial,
  default: tutorial,
};

export default files;
