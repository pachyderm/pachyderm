import {
  TaskComponentProps,
  TaskCard,
  useMinimizeTask,
} from '@pachyderm/components';
import React from 'react';

const MinimizeTask: React.FC<TaskComponentProps> = ({
  currentTask,
  onCompleted,
  minimized,
  index,
  name,
}) => {
  useMinimizeTask({currentTask, index, minimized, onCompleted});

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
