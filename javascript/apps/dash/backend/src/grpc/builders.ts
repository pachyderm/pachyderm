import {
  Commit,
  File,
  FileInfo,
  FileType,
  Repo,
} from '@pachyderm/proto/pb/pfs/pfs_pb';

export type FileObject = {
  commitId?: Commit.AsObject['id'];
  path?: File.AsObject['path'];
  repoName: Repo.AsObject['name'];
};

export type FileInfoObject = {
  file: FileObject;
  fileType: FileType;
  hash: FileInfo.AsObject['hash'];
  sizeBytes: FileInfo.AsObject['sizeBytes'];
};

export const fileFromObject = ({
  commitId = 'master',
  path = '/',
  repoName,
}: FileObject) => {
  const file = new File();
  const commit = new Commit();
  const repo = new Repo();

  repo.setName(repoName);
  commit.setId(commitId);
  commit.setRepo(repo);
  file.setPath(path);
  file.setCommit(commit);

  return file;
};

export const fileInfoFromObject = ({
  file,
  fileType,
  hash,
  sizeBytes,
}: FileInfoObject) => {
  const fileInfo = new FileInfo();

  fileInfo.setFile(fileFromObject(file));
  fileInfo.setFileType(fileType);
  fileInfo.setHash(hash);
  fileInfo.setSizeBytes(sizeBytes);

  return fileInfo;
};
