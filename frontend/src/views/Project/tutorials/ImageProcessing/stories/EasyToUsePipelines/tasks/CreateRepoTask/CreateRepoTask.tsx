import React from 'react';

import {useCreateRepoMutation} from '@dash-frontend/generated/hooks';
import useAccount from '@dash-frontend/hooks/useAccount';
import useRecordTutorialProgress from '@dash-frontend/hooks/useRecordTutorialProgress';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  LoadingDots,
  TaskCard,
  TaskComponentProps,
  Terminal,
} from '@pachyderm/components';

const CreateRepoTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  currentStory,
  index,
  name,
}) => {
  const {projectId} = useUrlState();
  const {tutorialId, loading: accountLoading} = useAccount();
  const recordTutorialProgress = useRecordTutorialProgress(
    'image-processing',
    currentStory,
    currentTask,
  );
  const [createRepo, {loading, error}] = useCreateRepoMutation({
    variables: {
      args: {
        name: `images_${tutorialId}`,
        projectId,
      },
    },
    onCompleted: () => {
      onCompleted();
      recordTutorialProgress();
    },
  });

  return (
    <TaskCard
      task={name}
      index={index}
      action={createRepo}
      error={error?.message}
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
