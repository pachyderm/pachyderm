import React from 'react';
import {closeIcon} from '@jupyterlab/ui-components';

type PipelineSplashProps = {
  setShowPipeline: (shouldShow: boolean) => void;
};

const PipelineSplash: React.FC<PipelineSplashProps> = ({setShowPipeline}) => {
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
      <div className="pachyderm-mount-pipeline-splash">
        <span>Open a notebook to create a pipeline</span>
      </div>
    </div>
  );
};

export default PipelineSplash;
