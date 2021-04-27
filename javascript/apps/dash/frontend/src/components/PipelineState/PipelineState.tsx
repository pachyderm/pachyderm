import {Circle, Group} from '@pachyderm/components';
import React, {useMemo} from 'react';

import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {PipelineState as PipelineStateEnum} from '@graphqlTypes';

interface PipelineStateProps {
  state: PipelineStateEnum;
}

const PipelineState: React.FC<PipelineStateProps> = ({state}) => {
  const color = useMemo(() => {
    switch (state) {
      case PipelineStateEnum.PIPELINE_PAUSED:
      case PipelineStateEnum.PIPELINE_RESTARTING:
      case PipelineStateEnum.PIPELINE_RUNNING:
      case PipelineStateEnum.PIPELINE_STANDBY:
      case PipelineStateEnum.PIPELINE_STARTING:
        return 'green';
      default:
        return 'red';
    }
  }, [state]);

  return (
    <Group spacing={8} align="center">
      <Circle color={color} />
      {readablePipelineState(state)}
    </Group>
  );
};

export default PipelineState;
