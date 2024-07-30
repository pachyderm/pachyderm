export type FileTreeProps = {
  selectedCommitId?: string;
  path?: string;
  open?: boolean;
  initial?: boolean;
  nextAfterDirectory?: string;
};

export type FileTreeNodeProps = {
  filePath?: string;
  fileType?: FileType;
  selectedCommitId: string;
  projectId: string;
  repoId: string;
  currentPath: string;
  open?: boolean;
  nextPath?: string;
  previousPath?: string;
};
