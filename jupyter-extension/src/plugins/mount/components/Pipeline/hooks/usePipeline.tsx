import {useEffect, useState} from 'react';
import {ServerConnection} from '@jupyterlab/services';

import {
  CreatePipelineResponse,
  GpuMode,
  MountSettings,
  PpsContext,
  PpsMetadata,
} from '../../../types';
import {requestAPI} from '../../../../../handler';

export const PPS_VERSION = 'v1.0.0';

export type usePipelineResponse = {
  loading: boolean;
  pipelineName: string;
  setPipelineName: (input: string) => void;
  pipelineProject: string;
  setPipelineProject: (input: string) => void;
  imageName: string;
  setImageName: (input: string) => void;
  inputSpec: string;
  setInputSpec: (input: string) => void;
  pipelinePort: string;
  setPipelinePort: (input: string) => void;
  gpuMode: GpuMode;
  setGpuMode: (input: GpuMode) => void;
  resourceSpec: string;
  setResourceSpec: (input: string) => void;
  requirements: string;
  setRequirements: (input: string) => void;
  callCreatePipeline: () => Promise<void>;
  currentNotebook: string;
  errorMessage: string;
  responseMessage: string;
};

export const usePipeline = (
  ppsContext: PpsContext | undefined,
  settings: MountSettings,
  saveNotebookMetaData: (metadata: PpsMetadata) => void,
  saveNotebookToDisk: () => Promise<string | null>,
): usePipelineResponse => {
  const [loading, setLoading] = useState(false);
  const [pipelineName, setPipelineName] = useState('');
  const [pipelineProject, setPipelineProject] = useState('');
  const [imageName, setImageName] = useState('');
  const [inputSpec, setInputSpec] = useState('');
  const [pipelinePort, setPipelinePort] = useState('');
  const [gpuMode, setGpuMode] = useState(GpuMode.None);
  const [resourceSpec, setResourceSpec] = useState('');
  const [requirements, setRequirements] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [responseMessage, setResponseMessage] = useState('');
  const [currentNotebook, setCurrentNotebook] = useState('None');

  useEffect(() => {
    setImageName(
      ppsContext?.metadata?.config.image ?? settings.defaultPipelineImage,
    );
    setPipelineName(ppsContext?.metadata?.config.pipeline.name ?? '');
    setPipelineProject(
      ppsContext?.metadata?.config.pipeline.project?.name ?? '',
    );
    setRequirements(ppsContext?.metadata?.config.requirements ?? '');
    setResponseMessage('');
    if (ppsContext?.metadata?.config.input_spec) {
      setInputSpec(ppsContext.metadata.config.input_spec);
    } else {
      setInputSpec('');
    }
    setCurrentNotebook(ppsContext?.notebookModel?.name ?? 'None');
    setPipelinePort(ppsContext?.metadata?.config.port ?? '');
    setGpuMode(ppsContext?.metadata?.config.gpu_mode ?? GpuMode.None);
    setResourceSpec(ppsContext?.metadata?.config.resource_spec ?? '');
  }, [ppsContext]);

  useEffect(() => {
    const ppsMetadata: PpsMetadata = buildMetadata();
    saveNotebookMetaData(ppsMetadata);
  }, [
    pipelineName,
    pipelineProject,
    imageName,
    requirements,
    inputSpec,
    pipelinePort,
    gpuMode,
    resourceSpec,
  ]);

  let callCreatePipeline: () => Promise<void>;
  if (ppsContext?.notebookModel) {
    const notebook = ppsContext.notebookModel;
    callCreatePipeline = async () => {
      setLoading(true);
      setErrorMessage('');
      setResponseMessage('');

      const ppsMetadata = buildMetadata();
      saveNotebookMetaData(ppsMetadata);
      const last_modified_time = await saveNotebookToDisk();

      try {
        const response = await requestAPI<CreatePipelineResponse>(
          `pps/_create/${encodeURI(notebook.path)}`,
          'PUT',
          {last_modified_time: last_modified_time ?? notebook.last_modified},
        );
        if (response.message !== null) {
          setResponseMessage(response.message);
        }
      } catch (e) {
        if (e instanceof ServerConnection.ResponseError) {
          // statusText is the only place that the user will get yaml parsing
          // errors (though it will also include Pachyderm errors, e.g.
          // "missing input repo").
          setErrorMessage('Error creating pipeline: ' + e.response.statusText);
        } else {
          throw e;
        }
      }
      setLoading(false);
    };
  } else {
    // If no notebookModel is defined, we cannot create a pipeline.
    callCreatePipeline = async () => {
      setErrorMessage('Error: No notebook in focus');
    };
  }

  const buildMetadata = (): PpsMetadata => {
    return {
      version: PPS_VERSION,
      config: {
        pipeline: {name: pipelineName, project: {name: pipelineProject}},
        image: imageName,
        requirements: requirements,
        input_spec: inputSpec,
        port: pipelinePort,
        gpu_mode: gpuMode,
        resource_spec: resourceSpec,
      },
    };
  };

  return {
    loading,
    pipelineName,
    setPipelineName,
    pipelineProject,
    setPipelineProject,
    imageName,
    setImageName,
    inputSpec,
    setInputSpec,
    pipelinePort,
    setPipelinePort,
    gpuMode,
    setGpuMode,
    resourceSpec,
    setResourceSpec,
    requirements,
    setRequirements,
    callCreatePipeline,
    currentNotebook,
    errorMessage,
    responseMessage,
  };
};
