import {useCallback, useEffect, useMemo, useRef, useState} from 'react';

import {useProgressBar} from '../../../../ProgressBar';
import {Story} from '../../../lib/types';

const useTutorialModal = (
  stories: Story[],
  initialStory: number,
  initialTask: number,
) => {
  const tutorialModalRef = useRef<HTMLDivElement>(null);
  const [currentTask, setCurrentTask] = useState(initialTask);
  const [currentStory, setCurrentStory] = useState(initialStory);
  const {visitStep, completeStep, clear} = useProgressBar();
  const taskSections = useMemo(
    () => stories[currentStory].sections.filter((section) => section.Task),
    [stories, currentStory],
  );

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

  const handleNextStory = () => {
    if (tutorialModalRef.current) tutorialModalRef.current.scrollTop = 0;

    setCurrentStory((prevValue) => Math.min(prevValue + 1, stories.length - 1));
    setCurrentTask(0);
    clear();
  };

  const handleStoryChange = useCallback(
    (name: string) => {
      if (name !== stories[currentStory].name) {
        if (tutorialModalRef.current) tutorialModalRef.current.scrollTop = 0;

        setCurrentStory(stories.findIndex((story) => story.name === name));
        setCurrentTask(0);
        clear();
      }
    },
    [stories, clear, currentStory],
  );

  return {
    currentStory,
    currentTask,
    handleNextStory,
    handleStoryChange,
    handleTaskCompletion,
    taskSections,
    tutorialModalRef,
  };
};

export default useTutorialModal;
