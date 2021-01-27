import {Node, PipelineState} from '@graphqlTypes';

const nodeStateAsPipelineState = (state: Node['state']) => {
  return PipelineState[(state as unknown) as keyof typeof PipelineState];
};

export default nodeStateAsPipelineState;
