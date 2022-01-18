import {TaskComponentProps, TaskCard, Terminal} from '@pachyderm/components';
import React, {useCallback} from 'react';

import useAccount from '@dash-frontend/hooks/useAccount';
import useCreateBranch from '@dash-frontend/hooks/useCreateBranch';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const MoveBranchTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  const {projectId} = useUrlState();
  const {tutorialId, loading: accountLoading} = useAccount();
  const onCreateBranch = useCallback(() => {
    onCompleted();
  }, [onCompleted]);
  const {createBranch} = useCreateBranch(onCreateBranch);
  const action = useCallback(() => {
    createBranch({
      head: {
        id: '^',
        branch: {name: 'master', repo: {name: `images_${tutorialId}`}},
      },
      branch: {name: 'master', repo: {name: `images_${tutorialId}`}},
      projectId,
    });
  }, [createBranch, projectId, tutorialId]);

  return (
    <TaskCard
      task={name}
      index={index}
      currentTask={currentTask}
      action={action}
      actionText="Move images branch"
      taskInfoTitle="Reproduce results through branch manipulation"
      taskInfo={
        <p>
          You&apos;ll move the master branch in the images repo to the very
          first commit in that repo, snapping the entire state of the DAG back
          to its former state.
        </p>
      }
      disabled={accountLoading}
    >
      <Terminal>
        {`pachctl create branch images_${tutorialId}@master --head images@master^`}
      </Terminal>
    </TaskCard>
  );
};

export default MoveBranchTask;
