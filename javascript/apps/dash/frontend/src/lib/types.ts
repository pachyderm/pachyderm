export interface ProjectRouteParams {
  projectId: string;
  repoId?: string;
  pipelineId?: string;
  branchId?: string;
  tabId?: string;
  commitId?: string;
  filePath?: string;
  jobId?: string;
  pipelineJobId?: string;
}

export type FileMajorType =
  | 'document'
  | 'image'
  | 'video'
  | 'audio'
  | 'folder'
  | 'unknown';
