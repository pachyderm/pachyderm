import {TutorialModal} from '@pachyderm/components';
import React from 'react';

import {TutorialProps} from '../ProjectTutorial';

import ExitSurvey from './ExitSurvey';
import useImageProcessing from './hooks/useImageProcessing';
import stories from './stories';

const ImageProcessing: React.FC<TutorialProps> = ({onClose}) => {
  const {exitSurveyOpen, onTutorialComplete} = useImageProcessing();
  return (
    <>
      {exitSurveyOpen ? <ExitSurvey onTutorialClose={onClose} /> : null}
      <TutorialModal
        onTutorialComplete={onTutorialComplete}
        stories={stories}
      />
    </>
  );
};

export default ImageProcessing;
