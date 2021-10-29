import {TaskComponentProps, TaskCard} from '@pachyderm/components';
import React, {useEffect} from 'react';

const MinimizeTask: React.FC<TaskComponentProps> = ({
  currentTask,
  onCompleted,
  minimized,
  index,
}) => {
  useEffect(() => {
    if (currentTask === index && minimized) {
      onCompleted();
    }
  }, [currentTask, minimized, onCompleted, index]);

  return (
    <>
      <TaskCard
        index={index}
        currentTask={currentTask}
        name="Minimize the overlay and inspect the pipeline and resulting output repo in the DAG"
      />
    </>
  );
};

export default MinimizeTask;
