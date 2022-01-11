import {TaskComponentProps, TaskCard, Terminal} from '@pachyderm/components';
import React, {useCallback} from 'react';

import useCreateBranch from '@dash-frontend/hooks/useCreateBranch';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const MoveBranchTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  const {projectId} = useUrlState();
  const onCreateBranch = useCallback(() => {
    onCompleted();
  }, [onCompleted]);
  const {createBranch} = useCreateBranch(onCreateBranch);
  const action = useCallback(() => {
    createBranch({
      head: {id: '^', branch: {name: 'master', repo: {name: 'images'}}},
      branch: {name: 'master', repo: {name: 'images'}},
      projectId,
    });
  }, [createBranch, projectId]);

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
    >
      <Terminal>
        pachctl create branch images@master --head images@master^
      </Terminal>
    </TaskCard>
  );
};

export default MoveBranchTask;
