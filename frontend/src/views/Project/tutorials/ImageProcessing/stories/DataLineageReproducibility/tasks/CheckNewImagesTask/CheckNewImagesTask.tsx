import React from 'react';

import useRecordTutorialProgress from '@dash-frontend/hooks/useRecordTutorialProgress';
import {
  TaskComponentProps,
  TaskCard,
  useMinimizeTask,
} from '@pachyderm/components';

const CheckNewImagesTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  currentStory,
  index,
  name,
  minimized,
}) => {
  const recordTutorialProgress = useRecordTutorialProgress(
    'image-processing',
    currentStory,
    currentTask,
  );
  useMinimizeTask({
    currentTask,
    index,
    minimized,
    onCompleted: () => {
      onCompleted();
      recordTutorialProgress();
    },
  });

  return (
    <TaskCard
      task={name}
      index={index}
      currentTask={currentTask}
      actionText="Minimize the overlay"
      taskInfoTitle="See the original montage file is restored"
      taskInfo={
        <p>
          Minimize this overlay and look at the head commit in the master branch
          of the montage repo. Examine the file. It should have a original
          montage..
        </p>
      }
    />
  );
};

export default CheckNewImagesTask;
