import {DiffFileResponse} from '@dash-frontend/api/pfs';
import formatBytes from '@dash-frontend/lib/formatBytes';
import {FileCommitState} from '@dash-frontend/lib/types';

const formatDiff = (diff: DiffFileResponse[]) => {
  const counts = {
    added: {count: 0, sizeDelta: 0},
    updated: {count: 0, sizeDelta: 0},
    deleted: {count: 0, sizeDelta: 0},
  };

  const diffTotals: Record<string, FileCommitState> = {};

  for (const fileDiff of diff) {
    if (fileDiff.newFile?.file && !fileDiff.oldFile?.file) {
      if (!fileDiff.newFile.file.path?.endsWith('/')) {
        counts.added.count += 1;
        counts.added.sizeDelta += Number(fileDiff.newFile.sizeBytes || 0);
      }
      diffTotals[fileDiff.newFile.file.path || ''] = FileCommitState.ADDED;
    } else if (!fileDiff.newFile?.file && fileDiff.oldFile?.file) {
      if (!fileDiff.oldFile.file.path?.endsWith('/')) {
        counts.deleted.count += 1;
        counts.deleted.sizeDelta -= Number(fileDiff.oldFile.sizeBytes || 0);
      }
      diffTotals[fileDiff.oldFile.file.path || ''] = FileCommitState.DELETED;
    } else if (fileDiff.newFile?.file && fileDiff.oldFile?.file) {
      if (!fileDiff.newFile.file.path?.endsWith('/')) {
        counts.updated.count += 1;
        counts.updated.sizeDelta +=
          (Number(fileDiff.newFile.sizeBytes) || 0) -
          Number(fileDiff.oldFile.sizeBytes || 0);
      }
      diffTotals[fileDiff.newFile.file.path || ''] = FileCommitState.UPDATED;
    }
  }

  const sizeDiff =
    (Number(diff[0]?.newFile?.sizeBytes) || 0) -
    (Number(diff[0]?.oldFile?.sizeBytes) || 0);

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

// This formatDiffOnlyTotals function serves as a temporary solution to improve the performance of the long-running synchronous functions above it,
// which block the main thread. It is a small optimization that aims to speed up their execution. If these functions ever get refactored to
// become asynchronous and non-blocking, the person responsible should consider revisiting and potentially merging these functions.
export const formatDiffOnlyTotals = (diff: DiffFileResponse[]) => {
  const diffTotals: Record<string, FileCommitState> = {};

  for (const fileDiff of diff) {
    if (fileDiff.newFile?.file && !fileDiff.oldFile?.file) {
      diffTotals[fileDiff.newFile.file.path || ''] = FileCommitState.ADDED;
    } else if (!fileDiff.newFile?.file && fileDiff.oldFile?.file) {
      diffTotals[fileDiff.oldFile.file.path || ''] = FileCommitState.DELETED;
    } else if (fileDiff.newFile?.file && fileDiff.oldFile?.file) {
      if (!fileDiff.newFile.file.path?.endsWith('/')) {
        diffTotals[fileDiff.newFile.file.path || ''] = FileCommitState.UPDATED;
      }
    }
  }
  return diffTotals;
};

export default formatDiff;
