import {
  TaskComponentProps,
  TaskCard,
  useMinimizeTask,
} from '@pachyderm/components';
import React from 'react';

import useRecordTutorialProgress from '@dash-frontend/hooks/useRecordTutorialProgress';

const MinimizeTask: React.FC<TaskComponentProps> = ({
  currentTask,
  currentStory,
  onCompleted,
  minimized,
  index,
  name,
}) => {
  const recordTutorialProgress = useRecordTutorialProgress(
    'image-processing',
    currentStory,
    currentTask,
    onCompleted,
  );

  useMinimizeTask({
    currentTask,
    index,
    minimized,
    onCompleted: recordTutorialProgress,
  });

  return (
    <TaskCard
      task={name}
      index={index}
      currentTask={currentTask}
      taskInfoTitle="Confirm skipped datums"
      taskInfo="Only the new files were processed. The original files were skipped."
    />
  );
};

export default MinimizeTask;
