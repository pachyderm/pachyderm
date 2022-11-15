import {useEffect} from 'react';

import {TaskComponentProps} from '../lib/types';

type useMinimizeTaskProps = Omit<TaskComponentProps, 'name' | 'currentStory'>;

const useMinimizeTask = ({
  currentTask,
  index,
  minimized,
  onCompleted,
}: useMinimizeTaskProps) => {
  useEffect(() => {
    if (currentTask === index && minimized) {
      onCompleted();
    }
  }, [currentTask, minimized, onCompleted, index]);
};

export default useMinimizeTask;
