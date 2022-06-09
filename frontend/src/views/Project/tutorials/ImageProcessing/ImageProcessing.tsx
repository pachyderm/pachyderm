import {TutorialModal} from '@pachyderm/components';
import React from 'react';

import {TutorialProps} from '../ProjectTutorial';

import ExitSurvey from './ExitSurvey';
import useImageProcessing from './hooks/useImageProcessing';
import stories from './stories';

const ImageProcessing: React.FC<TutorialProps> = ({onClose}) => {
  const {isExitSurveyOpen, openExitSurvey, closeTutorial, initialProgress} =
    useImageProcessing();

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
