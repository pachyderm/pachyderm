import {BranchesQueryArgs} from '@graphqlTypes';

import {useGetBranchesQuery} from '@dash-frontend/generated/hooks';

type UseBranchesArgs = {
  args: BranchesQueryArgs;
};

const useBranches = (args: UseBranchesArgs) => {
  const branchesQuery = useGetBranchesQuery({
    variables: args,
  });

  return {
    ...branchesQuery,
    branches: branchesQuery?.data?.branches,
  };
};

export default useBranches;
