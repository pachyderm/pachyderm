import React from 'react';

import ProgressBar from 'ProgressBar';

import TutorialModalBody from './components/TutorialModalBody';
import {Story} from './lib/types';

type TutorialModalProps = {
  stories: Story[];
  initialStep?: number;
  iniitalTask?: number;
};

const TutorialModal: React.FC<TutorialModalProps> = (props) => {
  return (
    <ProgressBar vertical>
      <TutorialModalBody {...props} />
    </ProgressBar>
  );
};

export default TutorialModal;
