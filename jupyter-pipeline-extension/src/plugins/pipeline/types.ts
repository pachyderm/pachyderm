import {Contents} from '@jupyterlab/services';



export type Project = {
  name: string;
};

export type ProjectAuthInfo = {
  permissions: number[];
  roles: string[];
};

export type ProjectInfo = {
  project: Project;
  auth: ProjectAuthInfo;
};

export type Pipeline = {
  name: string;
  project: Project | null;
};

export type PipelineSpec = {
  pipeline: Pipeline;
  description: string | null;
  transform: any;
  input: any;
  update: boolean;
  reprocess: boolean;
};

export type PpsMetadata = {
  version: string;
  config: PpsConfig;
};

// If this is updated, make sure to also update the corresponding `useEffect`
// call in ./components/Pipeline/hooks/usePipeline.tsx that writes this type to
// the notebook metadata.
export type PpsConfig = {
  pipeline: Pipeline;
  image: string;
  requirements: string | null;
  input_spec: string;
};

export type PpsContext = {
  metadata: PpsMetadata | null;
  notebookModel: Contents.IModel | null;
};

export type CreatePipelineResponse = {
  message: string | null;
};

export type PipelineSettings = { //Renamed from mount settings
  defaultPipelineImage: string;
};