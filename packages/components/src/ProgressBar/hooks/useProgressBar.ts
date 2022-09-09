import {useContext, useCallback} from 'react';

import ProgressBarContext from '../contexts/ProgressBarContext';

const useProgressBar = () => {
  const {visit, complete, visited, completed, isVertical, clear} =
    useContext(ProgressBarContext);

  const isVisited = useCallback(
    (id: string) => {
      return Boolean(visited.includes(id));
    },
    [visited],
  );

  const isCompleted = useCallback(
    (id: string) => {
      return Boolean(completed.includes(id));
    },
    [completed],
  );

  const visitStep = useCallback((id: string) => visit(id), [visit]);
  const completeStep = useCallback((id: string) => complete(id), [complete]);

  return {
    visitStep,
    completeStep,
    isVisited,
    isCompleted,
    isVertical,
    clear,
  };
};

export default useProgressBar;
