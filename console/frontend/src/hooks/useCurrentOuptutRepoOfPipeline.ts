import {useRepo} from './useRepo';
import useUrlState from './useUrlState';

const useCurrentOuptutRepoOfPipeline = () => {
  const {pipelineId, projectId} = useUrlState();
  const {repo, loading, error} = useRepo({
    repo: {
      name: pipelineId,
      project: {
        name: projectId,
      },
    },
  });

  return {
    repo,
    error,
    loading,
  };
};

export default useCurrentOuptutRepoOfPipeline;
