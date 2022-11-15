import React from 'react';

import ProgressBar from '@pachyderm/components/ProgressBar';

import TutorialModalBody from './components/TutorialModalBody';
import {TutorialModalBodyProps} from './components/TutorialModalBody/TutorialModalBody';

interface TutorialModalProps
  extends Pick<
    TutorialModalBodyProps,
    | 'stories'
    | 'initialTask'
    | 'initialStory'
    | 'onTutorialComplete'
    | 'onSkip'
    | 'onClose'
  > {
  initialStep?: number;
  tutorialName: string;
}

const TutorialModal: React.FC<TutorialModalProps> = (props) => {
  return (
    <ProgressBar vertical>
      <TutorialModalBody {...props} />
    </ProgressBar>
  );
};

export default TutorialModal;
