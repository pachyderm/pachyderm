import {FileCommitState} from '@dash-backend/generated/types';
import formatBytes from '@dash-backend/lib/formatBytes';
import {DiffFileResponse} from '@dash-backend/proto';

const formatDiff = (diff: DiffFileResponse.AsObject[]) => {
  let counts = {
    added: {count: 0, sizeDelta: 0},
    updated: {count: 0, sizeDelta: 0},
    deleted: {count: 0, sizeDelta: 0},
  };

  const diffTotals = diff.reduce<Record<string, FileCommitState>>(
    (acc, fileDiff) => {
      if (fileDiff.newFile?.file && !fileDiff.oldFile?.file) {
        if (!fileDiff.newFile.file.path.endsWith('/')) {
          counts = {
            ...counts,
            added: {
              count: counts.added.count + 1,
              sizeDelta: counts.added.sizeDelta + fileDiff.newFile?.sizeBytes,
            },
          };
        }
        return {
          ...acc,
          [fileDiff.newFile.file.path]: FileCommitState.ADDED,
        };
      }
      if (!fileDiff.newFile?.file && fileDiff.oldFile?.file) {
        if (!fileDiff.oldFile.file.path.endsWith('/')) {
          counts = {
            ...counts,
            deleted: {
              count: counts.deleted.count + 1,
              sizeDelta:
                counts.deleted.sizeDelta - (fileDiff.oldFile?.sizeBytes || 0),
            },
          };
        }
        return {
          ...acc,
          [fileDiff.oldFile.file.path]: FileCommitState.DELETED,
        };
      }
      if (fileDiff.newFile?.file && fileDiff.oldFile?.file) {
        if (!fileDiff.newFile.file.path.endsWith('/')) {
          counts = {
            ...counts,
            updated: {
              count: counts.updated.count + 1,
              sizeDelta:
                counts.updated.sizeDelta +
                fileDiff.newFile?.sizeBytes -
                fileDiff.oldFile?.sizeBytes,
            },
          };
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
