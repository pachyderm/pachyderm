import React from 'react';
import {closeIcon} from '@jupyterlab/ui-components';
import {usePipeline} from './hooks/usePipeline';

type PipelineProps = {
  showPipeline: boolean;
  setShowPipeline: (shouldShow: boolean) => void;
};

const placeholderInputSpec = `pfs:
  repo: images
  branch: dev
  glob: /*
`;
const placeholderRequirements = 'dependency=="1.2.3"';
const placeholderSecrets = 'MyPassword: ********';

const Pipeline: React.FC<PipelineProps> = ({showPipeline, setShowPipeline}) => {
  const {
    loading,
    inputSpec,
    setInputSpec,
    requirements,
    setRequirements,
    secrets,
    setSecrets,
    callCreatePipeline,
    errorMessage,
  } = usePipeline(showPipeline);

  return (
    <div className="pachyderm-mount-pipeline-base">
      <div className="pachyderm-mount-pipeline-back">
        <button
          data-testid="Datum__back"
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

      <button
        data-testid="Pipeline__create_pipeline"
        className="pachyderm-button-link"
        onClick={callCreatePipeline}
      >
        Create Pipeline
      </button>
      <span
        className="pachyderm-mount-pipeline-error"
        data-testid="Datum__errorMessage"
      >
        {errorMessage}
      </span>

      <div className="pachyderm-mount-pipeline-input-wrapper">
        <label className="pachyderm-mount-pipeline-label" htmlFor="inputSpec">
          Input Spec
        </label>
        <textarea
          className="pachyderm-input"
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

      <div className="pachyderm-mount-pipeline-input-wrapper">
        <label
          className="pachyderm-mount-pipeline-label"
          htmlFor="requirements"
        >
          requirements.txt
        </label>
        <textarea
          className="pachyderm-input"
          data-testid="Pipeline__inputSpecInput"
          name="requirements"
          value={requirements}
          onChange={(e: any) => {
            setRequirements(e.target.value);
          }}
          disabled={loading}
          placeholder={placeholderRequirements}
        ></textarea>
      </div>

      <div className="pachyderm-mount-pipeline-input-wrapper">
        <label className="pachyderm-mount-pipeline-label" htmlFor="secrets">
          Secrets
        </label>
        <textarea
          className="pachyderm-input"
          data-testid="Pipeline__inputSpecInput"
          name="secrets"
          value={secrets}
          onChange={(e: any) => {
            setSecrets(e.target.value);
          }}
          disabled={loading}
          placeholder={placeholderSecrets}
        ></textarea>
      </div>
    </div>
  );
};

export default Pipeline;
