import React from 'react';
import {closeIcon} from '@jupyterlab/ui-components';
import {usePipeline} from './hooks/usePipeline';
import {PpsContext, PpsMetadata, MountSettings} from '../../types';

type PipelineProps = {
  ppsContext: PpsContext | undefined;
  settings: MountSettings;
  setShowPipeline: (shouldShow: boolean) => void;
  saveNotebookMetadata: (metadata: PpsMetadata) => void;
  saveNotebookToDisk: () => Promise<string | null>;
};

const placeholderInputSpec = `# example:
pfs:
  repo: images
  branch: dev
  glob: /*
`;
const placeholderRequirements = './requirements.txt';

const Pipeline: React.FC<PipelineProps> = ({
  ppsContext,
  settings,
  setShowPipeline,
  saveNotebookMetadata,
  saveNotebookToDisk,
}) => {
  const {
    loading,
    pipeline,
    setPipeline,
    imageName,
    setImageName,
    inputSpec,
    setInputSpec,
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
          value={
            pipeline.project?.name
              ? `${pipeline.project.name}/${pipeline.name}`
              : pipeline.name
          }
          onChange={(e: any) => {
            setPipeline(e.target.value);
          }}
          disabled={loading}
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
  name: ${pipeline.name}
  project: ${pipeline.project?.name ?? 'default'}
transform:
  image: ${imageName}
input:
${inputSpec
  .split('\n')
  .map((line, _, __) => '  ' + line)
  .join('\n')}
`}
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
