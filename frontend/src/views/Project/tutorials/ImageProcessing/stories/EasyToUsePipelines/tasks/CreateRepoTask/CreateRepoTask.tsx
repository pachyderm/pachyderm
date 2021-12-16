import {
  LoadingDots,
  TaskCard,
  TaskComponentProps,
  Terminal,
} from '@pachyderm/components';
import React from 'react';

import {useCreateRepoMutation} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const CreateRepoTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  const {projectId} = useUrlState();
  const [createRepo, {loading}] = useCreateRepoMutation({
    variables: {
      args: {
        name: 'images',
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
    >
      {loading && <LoadingDots />}
      <Terminal>pachctl create repo images</Terminal>
    </TaskCard>
  );
};

export default CreateRepoTask;
