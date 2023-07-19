import React from 'react';
import {closeIcon} from '@jupyterlab/ui-components';
import {usePipeline} from './hooks/usePipeline';
import {PpsContext, PpsMetadata, MountSettings, GpuMode} from '../../types';

type PipelineProps = {
  ppsContext: PpsContext | undefined;
  settings: MountSettings;
  setShowPipeline: (shouldShow: boolean) => void;
  saveNotebookMetadata: (metadata: PpsMetadata) => void;
  saveNotebookToDisk: () => Promise<string | null>;
};

const placeholderAdvancedResourceSpec = `# example:
gpu:
  type: nvidia.com/gpu
  number: 1
`;
const placeholderSimpleResourceSpec = `gpu:
  type: nvidia.com/gpu
  number: 1
`;
const placeholderInputSpec = `# example:
pfs:
  repo: images
  branch: dev
  glob: /*
`;
const placeholderRequirements = './requirements.txt';
const placeholderProject = 'default';

const Pipeline: React.FC<PipelineProps> = ({
  ppsContext,
  settings,
  setShowPipeline,
  saveNotebookMetadata,
  saveNotebookToDisk,
}) => {
  const {
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
  } = usePipeline(
    ppsContext,
    settings,
    saveNotebookMetadata,
    saveNotebookToDisk,
  );

  return (
    <div className="pachyderm-mount-pipeline-base">
      <div className="pachyderm-mount-pipeline-back">
        <button
          data-testid="Pipeline__back"
          className="pachyderm-button-link"
          onClick={async () => {
            setShowPipeline(false);
          }}
        >
          Back{' '}
          <closeIcon.react
            tag="span"
            className="pachyderm-mount-icon-padding"
          />
        </button>
      </div>
      <span className="pachyderm-mount-pipeline-subheading">
        Publish as Pipeline
      </span>

      <div className="pachyderm-pipeline-current-notebook-wrapper">
        <label
          className="pachyderm-pipeline-current-notebook-label"
          htmlFor="currentNotebook"
        >
          Current Notebook:{'  '}
        </label>
        <span
          className="pachyderm-pipeline-current-notebook-value"
          data-testid="Pipeline__currentNotebookValue"
        >
          {currentNotebook}
        </span>
      </div>

      <div className="pachyderm-pipeline-input-wrapper">
        <label
          className="pachyderm-pipeline-input-label"
          htmlFor="pipelineName"
        >
          *Pipeline Name:{'  '}
        </label>
        <input
          className="pachyderm-pipeline-input"
          data-testid="Pipeline__inputPipelineName"
          name="pipelineName"
          value={pipelineName}
          onChange={(e: any) => {
            setPipelineName(e.target.value);
          }}
          disabled={loading}
        ></input>
      </div>
      <div className="pachyderm-pipeline-input-wrapper">
        <label
          className="pachyderm-pipeline-input-label"
          htmlFor="pipelineProjectName"
        >
          Pipeline Project Name:{'  '}
        </label>
        <input
          className="pachyderm-pipeline-input"
          data-testid="Pipeline__inputPipelineProjectName"
          name="pipelineName"
          value={pipelineProject}
          onChange={(e: any) => {
            setPipelineProject(e.target.value);
          }}
          disabled={loading}
          placeholder={placeholderProject}
        ></input>
      </div>
      <div className="pachyderm-pipeline-input-wrapper">
        <label className="pachyderm-pipeline-input-label" htmlFor="imageName">
          *Container Image Name:{'  '}
        </label>
        <input
          className="pachyderm-pipeline-input"
          data-testid="Pipeline__inputImageName"
          name="imageName"
          value={imageName}
          onChange={(e: any) => {
            setImageName(e.target.value);
          }}
          disabled={loading}
        ></input>
      </div>
      <div className="pachyderm-pipeline-input-wrapper">
        <label
          className="pachyderm-pipeline-input-label"
          htmlFor="requirements"
        >
          Requirements File:{'  '}
        </label>
        <input
          className="pachyderm-pipeline-input"
          data-testid="Pipeline__inputRequirements"
          name="requirements"
          value={requirements}
          onChange={(e: any) => {
            setRequirements(e.target.value);
          }}
          disabled={loading}
          placeholder={placeholderRequirements}
        ></input>
      </div>
      <div className="pachyderm-pipeline-input-wrapper">
        <label className="pachyderm-pipeline-input-label" htmlFor="port">
          Port:{'  '}
        </label>
        <input
          className="pachyderm-pipeline-input"
          data-testid="Pipeline__inputPort"
          name="port"
          value={pipelinePort}
          onChange={(e: any) => {
            setPipelinePort(e.target.value);
          }}
          disabled={loading}
        ></input>
      </div>
      <div className="pachyderm-pipeline-textarea-wrapper">
        <label
          className="pachyderm-pipeline-textarea-label"
          htmlFor="inputSpec"
        >
          Pipeline Input Spec:
        </label>
        <textarea
          className="pachyderm-pipeline-textarea pachyderm-input"
          data-testid="Pipeline__inputSpecInput"
          name="inputSpec"
          value={inputSpec}
          onChange={(e: any) => {
            setInputSpec(e.target.value);
          }}
          disabled={loading}
          placeholder={placeholderInputSpec}
        ></textarea>
      </div>
      <div className="pachyderm-pipeline-input-wrapper">
        <label className="pachyderm-pipeline-input-label" htmlFor="gpuMode">
          *Gpu Mode:{'  '}
        </label>
        <select
          className="pachyderm-pipeline-input"
          data-testid="Pipeline__gpuMode"
          name="gpuMode"
          value={gpuMode}
          onChange={(e: any) => {
            setGpuMode(e.target.value);
          }}
        >
          {(Object.keys(GpuMode) as Array<GpuMode>).map((mode) => (
            <option value={mode}>{mode}</option>
          ))}
        </select>
      </div>
      {gpuMode === GpuMode.Advanced && (
        <div className="pachyderm-pipeline-textarea-wrapper">
          <label
            className="pachyderm-pipeline-textarea-label"
            htmlFor="resourceSpec"
          >
            Pipeline Resource Spec:
          </label>
          <textarea
            className="pachyderm-pipeline-textarea pachyderm-input"
            data-testid="Pipeline__resourceSpecInput"
            name="resourceSpec"
            value={resourceSpec}
            onChange={(e: any) => {
              setResourceSpec(e.target.value);
            }}
            disabled={loading}
            placeholder={placeholderAdvancedResourceSpec}
          ></textarea>
        </div>
      )}
      <div className="pachyderm-pipeline-spec-preview pachyderm-pipeline-textarea-wrapper">
        <label className="pachyderm-pipeline-preview-label">
          Pipeline Spec Preview: {'  '}
        </label>
        <textarea
          className="pachyderm-pipeline-spec-preview-textarea"
          style={{backgroundColor: '#80808080'}}
          data-testid="Pipeline__specPreview"
          name="specPreview"
          value={`pipeline:
  name: ${pipelineName}
  project: ${pipelineProject || placeholderProject}
transform:
  image: ${imageName}
input:
${inputSpec
  .split('\n')
  .map((line, _, __) => '  ' + line)
  .join('\n')}
${
  gpuMode === GpuMode.None
    ? ''
    : 'resource_limits:\n' +
      (gpuMode === GpuMode.Simple
        ? placeholderSimpleResourceSpec
        : gpuMode === GpuMode.Advanced
        ? resourceSpec
        : ''
      )
        .split('\n')
        .map((line, _, __) => '  ' + line)
        .join('\n')
}`}
          readOnly={true}
        ></textarea>
      </div>

      <div className="pachyderm-pipeline-buttons">
        <button
          data-testid="Pipeline__create_pipeline"
          className="pachyderm-button"
          onClick={callCreatePipeline}
        >
          Run
        </button>
      </div>
      <span
        className="pachyderm-pipeline-error"
        data-testid="Pipeline__errorMessage"
      >
        {errorMessage}
      </span>
      <span
        className="pachyderm-pipeline-response"
        data-testid="Pipeline__responseMessage"
      >
        {responseMessage}
      </span>
    </div>
  );
};

export default Pipeline;
