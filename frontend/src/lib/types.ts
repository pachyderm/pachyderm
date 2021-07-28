import {CSSProperties} from 'react';

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

export type SidebarSize = 'sm' | 'md' | 'lg';

export type FixedListRowProps = {
  index: number;
  style: CSSProperties;
};

export type FixedGridRowProps = {
  columnIndex: number;
  rowIndex: number;
  style: CSSProperties;
};
