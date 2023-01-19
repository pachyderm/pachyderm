import {FileType, DiffFileResponse} from '@dash-backend/proto';
import {fileInfoFromObject} from '@dash-backend/proto/builders/pfs';

const diffResponseUpdated = new DiffFileResponse();
diffResponseUpdated.setNewFile(
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
);
diffResponseUpdated.setOldFile(
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
);

const diffResponseAdded = new DiffFileResponse();
diffResponseAdded.setNewFile(
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
);

const emptyDiffResponse = {'/': new DiffFileResponse()};

const tutorial = {
  '/': diffResponseUpdated,
};

const customer = {
  '/': diffResponseAdded,
};

export type Diffs = {
  [projectId: string]: {
    [path: string]: DiffFileResponse;
  };
};

const files: Diffs = {
  'Solar-Panel-Data-Sorting': customer,
  'Data-Cleaning-Process': tutorial,
  'Solar-Power-Data-Logger-Team-Collab': customer,
  'Solar-Price-Prediction-Modal': emptyDiffResponse,
  'Egress-Examples': emptyDiffResponse,
  'Empty-Project': emptyDiffResponse,
  'Trait-Discovery': emptyDiffResponse,
  'OpenCV-Tutorial': emptyDiffResponse,
  default: emptyDiffResponse,
};

export default files;
