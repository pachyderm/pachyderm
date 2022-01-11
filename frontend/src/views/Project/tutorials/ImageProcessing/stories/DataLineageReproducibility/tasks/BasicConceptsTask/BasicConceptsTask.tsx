import {
  TaskComponentProps,
  TaskCard,
  useMinimizeTask,
} from '@pachyderm/components';
import React from 'react';

const BasicConceptsTask: React.FC<TaskComponentProps> = ({
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
          of the montage repo. Examine the file. It should have a montage of all
          the current images in the images repo alongside their edge detected
          versions.
        </p>
      }
    />
  );
};

export default BasicConceptsTask;
