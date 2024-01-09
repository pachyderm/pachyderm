import {FileInfo, FileType} from '@dash-frontend/generated/proto/pfs/pfs.pb';

export const FILE_DOWNLOAD_LIMIT = 2e8; // 200 MB

export const getDownloadLink = (file: FileInfo) => {
  if (
    file.fileType !== FileType.DIR &&
    (Number(file.sizeBytes) || 0) <= FILE_DOWNLOAD_LIMIT
  ) {
    const repoName = file.file?.commit?.branch?.repo?.name;
    const branchName = file.file?.commit?.branch?.name;
    const commitId = file.file?.commit?.id;
    const filePath = file.file?.path;
    const projectId = file.file?.commit?.branch?.repo?.project?.name;

    if (repoName && branchName && commitId && filePath) {
      return `/proxyForward/pfs/${projectId}/${repoName}/${commitId}${filePath}`;
    }
  }
  return null;
};
