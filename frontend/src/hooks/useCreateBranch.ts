import {CreateBranchArgs} from '@graphqlTypes';

import {useCreateBranchMutation} from '@dash-frontend/generated/hooks';

const useCreateBranch = (onCompleted?: () => void) => {
  const [createBranchMutation, status] = useCreateBranchMutation({onCompleted});
  return {
    createBranch: (args: CreateBranchArgs) =>
      createBranchMutation({variables: {args}}),
    status,
  };
};

export default useCreateBranch;
