import {Circle, Group} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import React, {useMemo} from 'react';

import pipelineStateAsString from '@dash-frontend/lib/pipelineStateAsString';
import {PipelineState as PipelineStateEnum} from '@graphqlTypes';

interface PipelineStateProps {
  state: PipelineStateEnum;
}

const PipelineState: React.FC<PipelineStateProps> = ({state}) => {
  const color = useMemo(() => {
    const enumValue =
      PipelineStateEnum[(state as unknown) as keyof typeof PipelineStateEnum];

    switch (enumValue) {
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
      {capitalize(pipelineStateAsString(state))}
    </Group>
  );
};

export default PipelineState;
