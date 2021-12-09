import {TutorialModal} from '@pachyderm/components';
import React from 'react';

import stories from './stories';

const OpenCvTutorial = () => {
  return <TutorialModal stories={stories} />;
};

export default OpenCvTutorial;
