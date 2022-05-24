import {Button, EducationSVG} from '@pachyderm/components';
import React from 'react';

import useRunTutorialButton from './hooks/useRunTutorialButton';

const RunTutorialButton = ({projectId}: {projectId?: string}) => {
  const {activeTutorial, startTutorial, tutorialProgress} =
    useRunTutorialButton(projectId);

  if (activeTutorial) {
    return null;
  }

  return (
    <Button
      buttonType="tertiary"
      IconSVG={EducationSVG}
      onClick={startTutorial}
    >
      {tutorialProgress ? 'Resume Tutorial' : 'Run Tutorial'}
    </Button>
  );
};

export default RunTutorialButton;
