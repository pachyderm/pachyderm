import React from 'react';
import {closeIcon} from '@jupyterlab/ui-components';
import {usePipeline} from './hooks/usePipeline';
import {SameMetadata} from '../../types';

type PipelineProps = {
  setShowPipeline: (shouldShow: boolean) => void;
  notebookPath: string | undefined;
  saveNotebookMetadata: (metadata: any) => void;
  metadata: SameMetadata | undefined;
};

const placeholderInputSpec = `pfs:
  repo: images
  branch: dev
  glob: /*
`;
const placeholderRequirements = './requirements.txt';

const Pipeline: React.FC<PipelineProps> = ({
  setShowPipeline,
  notebookPath,
  saveNotebookMetadata,
  metadata,
}) => {
  const {
    loading,
    pipelineName,
    setPipelineName,
    imageName,
    setImageName,
    inputSpec,
    setInputSpec,
    requirements,
    setRequirements,
    callCreatePipeline,
    callSavePipeline,
    errorMessage,
  } = usePipeline(metadata, notebookPath, saveNotebookMetadata);

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
        Notebook-to-Pipeline
      </span>

      <div className="pachyderm-pipeline-buttons">
        <button
          data-testid="Pipeline__save"
          className="pachyderm-button-link"
          onClick={callSavePipeline}
        >
          Save
        </button>

        <button
          data-testid="Pipeline__create_pipeline"
          className="pachyderm-button-link"
          onClick={callCreatePipeline}
        >
          Create Pipeline
        </button>
      </div>

      <span
        className="pachyderm-pipeline-error"
        data-testid="Pipeline__errorMessage"
      >
        {errorMessage}
      </span>

      <div className="pachyderm-pipeline-input-wrapper">
        <label
          className="pachyderm-pipeline-input-label"
          htmlFor="pipelineName"
        >
          *Name:{'  '}
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
        <label className="pachyderm-pipeline-input-label" htmlFor="imageName">
          *Image:{'  '}
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
          Requirements:{'  '}
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
          Input Spec
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
          value={`name: ${pipelineName}
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
    </div>
  );
};

export default Pipeline;
