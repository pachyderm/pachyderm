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
      actionText="Minimize this overlay"
      taskInfoTitle="Global identifiers tie together code and data"
      taskInfo={
        <>
          <p>
            Minimize this overlay and view the default project in the Console.
            Select <q>Show Jobs</q> from the top bar. Each pipeline is listed
            along the right. Selecting each pipeline will show its output commit
            for this job, where you can see the files associated with that job
            and commit. You can also see the transform used to produce that
            commit by scrolling down to transform for each pipeline.
          </p>
          <p>
            Pachyderm maintains versions of all data produced and automatically
            associates it with your inputs, saving you time.
          </p>
        </>
      }
    />
  );
};

export default MinimizeTask;
