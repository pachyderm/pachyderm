export interface ProjectRouteParams {
  projectId: string;
  repoId?: string;
  pipelineId?: string;
  branchId?: string;
  tabId?: string;
  commitId?: string;
  filePath?: string;
}

export type FileMajorType =
  | 'document'
  | 'image'
  | 'video'
  | 'audio'
  | 'folder'
  | 'unknown';
