import {TutorialModal} from '@pachyderm/components';
import React from 'react';

import useTutorialsList from '../hooks/useTutorialsList';
import {TutorialProps} from '../ProjectTutorial';

import ExitSurvey from './ExitSurvey';
import useImageProcessing from './hooks/useImageProcessing';
import stories from './stories';

const ImageProcessing: React.FC<TutorialProps> = ({onClose}) => {
  const {isExitSurveyOpen, openExitSurvey, closeTutorial} =
    useImageProcessing();
  const tutorials = useTutorialsList();
  const initialProgress = tutorials['image-processing'].progress;

  return (
    <>
      {isExitSurveyOpen ? <ExitSurvey onTutorialClose={onClose} /> : null}
      <TutorialModal
        onTutorialComplete={openExitSurvey}
        tutorialName="image-processing"
        stories={stories}
        onClose={closeTutorial}
        initialStory={initialProgress.initialStory}
        initialTask={initialProgress.initialTask}
      />
    </>
  );
};

export default ImageProcessing;
