import {CommitInfo, RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

const getSizeBytes = (obj: CommitInfo.AsObject | RepoInfo.AsObject) => {
  return obj.details?.sizeBytes || obj.sizeBytesUpperBound || 0;
};

export default getSizeBytes;
