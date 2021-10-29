import {TutorialModal} from '@pachyderm/components';
import React from 'react';

import steps from './steps';

const OpenCvTutorial = () => {
  return <TutorialModal steps={steps} />;
};

export default OpenCvTutorial;
