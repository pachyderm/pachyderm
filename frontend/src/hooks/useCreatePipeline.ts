import {CreatePipelineArgs} from '@graphqlTypes';

import {useCreatePipelineMutation} from '@dash-frontend/generated/hooks';

const useCreatePipeline = (
  args: CreatePipelineArgs,
  onCompleted: () => void,
) => {
  const [createPipelineMutation, status] = useCreatePipelineMutation({
    variables: {args},
    onCompleted,
  });

  return {
    createPipeline: createPipelineMutation,
    status,
  };
};

export default useCreatePipeline;
