import {
  TaskComponentProps,
  TaskCard,
  useMinimizeTask,
} from '@pachyderm/components';
import React from 'react';

const CheckMontageVersionTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
  minimized,
}) => {
  useMinimizeTask({currentTask, index, minimized, onCompleted});
  return (
    <TaskCard
      task={name}
      index={index}
      currentTask={currentTask}
      actionText="Minimize the overlay"
      taskInfoTitle="See the current montage file"
      taskInfo={
        <p>
          Minimize this overlay and look at the head commit in the master branch
          of the montage repo. Examine the file. It should have a new montage
          with the new images in addition to the old ones.
        </p>
      }
    />
  );
};

export default CheckMontageVersionTask;
