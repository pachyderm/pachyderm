import {useCallback, useState} from 'react';

import {Step} from 'TutorialModal/lib/types';

const useTutorialModal = (
  steps: Step[],
  initialStep: number,
  initialTask: number,
) => {
  const [minimized, setMinimized] = useState(false);
  const [currentTask, setCurrentTask] = useState(initialTask);
  const [currentStep, setCurrentStep] = useState(initialStep);

  const handleTaskCompletion = useCallback(
    (index: number) => {
      if (currentTask === index) setCurrentTask((prevValue) => prevValue + 1);
    },
    [currentTask],
  );

  const handleNextStep = () => {
    setCurrentStep((prevValue) => Math.min(prevValue + 1, steps.length - 1));
    setCurrentTask(0);
  };

  const displayTaskIndex = currentTask === 0 ? 0 : currentTask - 1;
  const displayTaskInstance =
    steps[currentStep].sections[displayTaskIndex].taskName;
  const nextTaskIndex = displayTaskIndex === 0 ? 1 : displayTaskIndex + 1;

  return {
    currentStep,
    currentTask,
    displayTaskIndex,
    displayTaskInstance,
    handleNextStep,
    handleTaskCompletion,
    minimized,
    nextTaskIndex,
    setCurrentStep,
    setMinimized,
  };
};

export default useTutorialModal;
