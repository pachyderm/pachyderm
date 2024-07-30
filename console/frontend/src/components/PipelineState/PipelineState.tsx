import React, {useMemo} from 'react';

import {PipelineState as PipelineStateEnum} from '@dash-frontend/api/pps';
import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {Circle, Group} from '@pachyderm/components';

interface PipelineStateProps {
  state?: PipelineStateEnum;
}

const PipelineState: React.FC<PipelineStateProps> = ({state}) => {
  const color = useMemo(() => {
    switch (state) {
      case PipelineStateEnum.PIPELINE_RESTARTING:
      case PipelineStateEnum.PIPELINE_RUNNING:
      case PipelineStateEnum.PIPELINE_STANDBY:
      case PipelineStateEnum.PIPELINE_STARTING:
        return 'green';
      case PipelineStateEnum.PIPELINE_PAUSED:
        return 'yellow';
      default:
        return 'red';
    }
  }, [state]);

  return (
    <Group spacing={8} align="center">
      <Circle color={color} />
      <h6>{readablePipelineState(state)}</h6>
    </Group>
  );
};

export default PipelineState;
