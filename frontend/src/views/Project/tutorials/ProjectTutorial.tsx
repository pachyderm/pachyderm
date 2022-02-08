import React, {useCallback, useEffect} from 'react';
import {Prompt} from 'react-router';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import ImageProcessing from './ImageProcessing';

export type TutorialProps = {
  onClose: () => void;
};

type TutorialMap = {
  [key: string]: React.FC<TutorialProps>;
};

const TUTORIALS: TutorialMap = {
  'image-processing': ImageProcessing,
};

const WARNING_MESSAGE =
  'Are you sure you want to leave? All tutorial progress will be lost.';

const ProjectTutorial: React.FC = () => {
  const {projectId} = useUrlState();
  const {viewState, updateViewState} = useUrlQueryState();

  const [id, setActiveTutorial] = useLocalProjectSettings({
    projectId,
    key: 'active_tutorial',
  });

  const onClose = useCallback(() => {
    updateViewState({tutorialId: undefined});
    setActiveTutorial(null);
  }, [updateViewState, setActiveTutorial]);

  useEffect(() => {
    if (viewState.tutorialId && viewState.tutorialId !== id) {
      setActiveTutorial(viewState.tutorialId);
    }
  }, [viewState.tutorialId, setActiveTutorial, id]);

  useEffect(() => {
    if (id) {
      const beforeUnloadListener = (event: BeforeUnloadEvent) => {
        event.preventDefault();

        // Firefox will automatically trigger the unload prompt by
        // having a truthy event listener. Chrome requires you to have a
        // non-undefined returned value.
        event.returnValue = '';
      };

      window.addEventListener('beforeunload', beforeUnloadListener, {
        capture: true,
      });

      window.addEventListener('unload', onClose);

      return () => {
        window.removeEventListener('beforeunload', beforeUnloadListener, {
          capture: true,
        });

        window.removeEventListener('unload', onClose);

        onClose();
      };
    }
  }, [id, onClose]);

  if (id) {
    const Tutorial = TUTORIALS[id];
    if (Tutorial) {
      return (
        <>
          <Prompt message={WARNING_MESSAGE} />
          <Tutorial onClose={onClose} />
        </>
      );
    }
  }
  return null;
};

export default ProjectTutorial;
