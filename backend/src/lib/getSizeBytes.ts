import {CommitInfo, RepoInfo} from '@dash-backend/proto';

const getSizeBytes = (obj: CommitInfo.AsObject | RepoInfo.AsObject) => {
  return obj.details?.sizeBytes || obj.sizeBytesUpperBound || 0;
};

export default getSizeBytes;
