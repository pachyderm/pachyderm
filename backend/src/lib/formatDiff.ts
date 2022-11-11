import {FileCommitState} from '@dash-backend/generated/types';
import formatBytes from '@dash-backend/lib/formatBytes';
import {DiffFileResponse} from '@dash-backend/proto';

const formatDiff = (diff: DiffFileResponse.AsObject[]) => {
  let counts = {added: 0, updated: 0, deleted: 0};

  const diffTotals = diff.reduce<Record<string, FileCommitState>>(
    (acc, fileDiff) => {
      if (fileDiff.newFile?.file && !fileDiff.oldFile?.file) {
        if (!fileDiff.newFile.file.path.endsWith('/')) {
          counts = {...counts, added: counts.added + 1};
        }
        return {
          ...acc,
          [fileDiff.newFile.file.path]: FileCommitState.ADDED,
        };
      }
      if (!fileDiff.newFile?.file && fileDiff.oldFile?.file) {
        if (!fileDiff.oldFile.file.path.endsWith('/')) {
          counts = {...counts, deleted: counts.deleted + 1};
        }
        return {
          ...acc,
          [fileDiff.oldFile.file.path]: FileCommitState.DELETED,
        };
      }
      if (fileDiff.newFile?.file && fileDiff.oldFile?.file) {
        if (!fileDiff.newFile.file.path.endsWith('/')) {
          counts = {...counts, updated: counts.updated + 1};
        }
        return {
          ...acc,
          [fileDiff.newFile.file.path]: FileCommitState.UPDATED,
        };
      }
      return acc;
    },
    {},
  );

  const sizeDiff =
    (diff[0]?.newFile?.sizeBytes || 0) - (diff[0]?.oldFile?.sizeBytes || 0);

  return {
    diffTotals,
    diff: {
      size: sizeDiff,
      sizeDisplay: formatBytes(sizeDiff),
      filesAdded: counts.added,
      filesUpdated: counts.updated,
      filesDeleted: counts.deleted,
    },
  };
};

export default formatDiff;
