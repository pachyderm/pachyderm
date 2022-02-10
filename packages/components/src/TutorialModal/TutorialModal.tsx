import React from 'react';

import ProgressBar from 'ProgressBar';

import TutorialModalBody from './components/TutorialModalBody';
import {TutorialModalBodyProps} from './components/TutorialModalBody/TutorialModalBody';

interface TutorialModalProps
  extends Pick<
    TutorialModalBodyProps,
    'stories' | 'initialTask' | 'onTutorialComplete' | 'onSkip'
  > {
  initialStep?: number;
}

const TutorialModal: React.FC<TutorialModalProps> = (props) => {
  return (
    <ProgressBar vertical>
      <TutorialModalBody {...props} />
    </ProgressBar>
  );
};

export default TutorialModal;
