import {TutorialModal} from '@pachyderm/components';
import React from 'react';

import useTutorialsList from '../hooks/useTutorialsList';
import {TutorialProps} from '../ProjectTutorial';

import useImageProcessing from './hooks/useImageProcessing';
import stories from './stories';

const ImageProcessing: React.FC<TutorialProps> = ({onClose}) => {
  const {closeTutorial} = useImageProcessing();
  const tutorials = useTutorialsList();
  const initialProgress = tutorials['image-processing'].progress;

  return (
    <>
      <TutorialModal
        onTutorialComplete={onClose}
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
