import {
  LoadingDots,
  TaskCard,
  TaskComponentProps,
  Terminal,
} from '@pachyderm/components';
import React from 'react';

import {useCreateRepoMutation} from '@dash-frontend/generated/hooks';
import useAccount from '@dash-frontend/hooks/useAccount';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const CreateRepoTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  const {projectId} = useUrlState();
  const {tutorialId, loading: accountLoading} = useAccount();
  const [createRepo, {loading}] = useCreateRepoMutation({
    variables: {
      args: {
        name: `images_${tutorialId}`,
        projectId,
      },
    },
    onCompleted,
  });

  return (
    <TaskCard
      task={name}
      index={index}
      action={createRepo}
      currentTask={currentTask}
      actionText="Create the images repo"
      taskInfoTitle="Create the images repo"
      taskInfo={
        'You must create the images repo before you create the "edges" pipeline that uses it'
      }
      disabled={accountLoading}
    >
      {loading && <LoadingDots />}
      <Terminal>{`pachctl create repo images_${tutorialId}`}</Terminal>
    </TaskCard>
  );
};

export default CreateRepoTask;
