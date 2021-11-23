import {useCallback, useEffect, useState} from 'react';

import {Story} from 'TutorialModal/lib/types';

import {useProgressBar} from '../../../../ProgressBar';

const useTutorialModal = (
  stories: Story[],
  initialStory: number,
  initialTask: number,
) => {
  const [minimized, setMinimized] = useState(false);
  const [currentTask, setCurrentTask] = useState(initialTask);
  const [currentStory, setCurrentStory] = useState(initialStory);
  const {visitStep, completeStep, clear} = useProgressBar();

  useEffect(() => {
    visitStep('0');
  });

  const handleTaskCompletion = useCallback(
    (index: number) => {
      if (currentTask === index) {
        completeStep(currentTask.toString());
        visitStep((currentTask + 1).toString());
        setCurrentTask((prevValue) => {
          return prevValue + 1;
        });
      }
    },
    [currentTask, completeStep, visitStep],
  );

  const handleNextStep = () => {
    setCurrentStory((prevValue) => Math.min(prevValue + 1, stories.length - 1));
    setCurrentTask(0);
    clear();
  };

  const handleStoryChange = useCallback(
    (name: string) => {
      if (name !== stories[currentStory].name) {
        setCurrentStory(stories.findIndex((step) => step.name === name));
        setCurrentTask(0);
        clear();
      }
    },
    [stories, clear, currentStory],
  );

  const displayTaskIndex = currentTask === 0 ? 0 : currentTask - 1;
  const displayTaskInstance =
    stories[currentStory].sections[displayTaskIndex].taskName;
  const nextTaskIndex = displayTaskIndex === 0 ? 1 : displayTaskIndex + 1;

  return {
    currentStory,
    currentTask,
    displayTaskIndex,
    displayTaskInstance,
    handleNextStep,
    handleStoryChange,
    handleTaskCompletion,
    minimized,
    nextTaskIndex,
    setCurrentStory,
    setMinimized,
  };
};

export default useTutorialModal;
