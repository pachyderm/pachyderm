import {TaskComponentProps, TaskCard} from '@pachyderm/components';
import React, {useEffect} from 'react';

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
    onCompleted,
  );

  useEffect(() => {
    if (minimized && currentTask === index) {
      recordTutorialProgress();
    }
  }, [minimized, currentTask, index, onCompleted, recordTutorialProgress]);

  return (
    <TaskCard
      task={name}
      index={index}
      currentTask={currentTask}
      actionText="Minimize this overlay"
      taskInfoTitle="Pachyderm automatically versions your data"
      taskInfo={
        <>
          <p>
            Minimize this overlay and view the default project in the Console.
            Select the montage repository and click &quot;View File&quot; in the
            top commit in the master branch. You can view all the files that
            have been processed. Click the X in the upper right corner to close
            it. Select the edges repo, which has a cube icon like the images
            repo, and perform the same action to see the previously processed
            data.
          </p>
        </>
      }
    />
  );
};

export default MinimizeTask;
