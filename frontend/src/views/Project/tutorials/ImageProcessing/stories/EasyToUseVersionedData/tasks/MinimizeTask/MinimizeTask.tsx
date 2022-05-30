import {
  TaskCard,
  TaskComponentProps,
  useMinimizeTask,
} from '@pachyderm/components';
import React from 'react';

import useRecordTutorialProgress from '@dash-frontend/hooks/useRecordTutorialProgress';

const MinimizeTask: React.FC<TaskComponentProps> = ({
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
      actionText="Minimize this overlay"
      taskInfoTitle="Pachyderm automatically versions your data"
      taskInfo={
        <p>
          Minimize this overlay and view the default project in the Console.
          Select the images repository and click <q>View Files</q> in the top
          commit in the master branch. You can view all the files that
          you&apos;ve input. Click the X in the upper right corner to close it.
          Select the edges repo, which has a cube icon like the images repo, and
          perform the same action to see the processed data.
        </p>
      }
    />
  );
};

export default MinimizeTask;
