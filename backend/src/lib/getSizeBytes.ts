import {CommitInfo, RepoInfo} from '@pachyderm/node-pachyderm';

const getSizeBytes = (obj: CommitInfo.AsObject | RepoInfo.AsObject) => {
  return obj.details?.sizeBytes || obj.sizeBytesUpperBound || 0;
};

export default getSizeBytes;
