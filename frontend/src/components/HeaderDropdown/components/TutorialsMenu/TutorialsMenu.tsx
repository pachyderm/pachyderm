import React, {useRef, useCallback, useContext} from 'react';

import useTutorialsList from '@dash-frontend/views/Project/tutorials/hooks/useTutorialsList';
import {useOutsideClick, TutorialModalBodyContext} from '@pachyderm/components';

import TutorialItem from './components/TutorialItem';
import useTutorialsMenu from './hooks/useTutorialsMenu';
import styles from './TutorialsMenu.module.css';

type TutorialsMenuProps = {
  projectId?: string;
  setTutorialsMenu: React.Dispatch<React.SetStateAction<boolean>>;
  stickTutorialsMenu: boolean;
  setStickTutorialsMenu: React.Dispatch<React.SetStateAction<boolean>>;
};

const TutorialsMenu: React.FC<TutorialsMenuProps> = ({
  projectId,
  setTutorialsMenu,
  stickTutorialsMenu,
  setStickTutorialsMenu,
}) => {
  const menuRef = useRef<HTMLDivElement>(null);
  const tutorialsList = useTutorialsList();
  const tutorials = Object.values(tutorialsList);
  const {startTutorial, tutorialProgress, deleteTutorialResources} =
    useTutorialsMenu(projectId);
  const handleOutsideClick = useCallback(() => {
    if (!stickTutorialsMenu) {
      setTutorialsMenu(false);
    }
  }, [setTutorialsMenu, stickTutorialsMenu]);
  useOutsideClick(menuRef, handleOutsideClick);
  const {setMinimized: setTutorialMinimized} = useContext(
    TutorialModalBodyContext,
  );

  const getHandleClick = (tutorialId: string) => () => {
    startTutorial(tutorialId);
    setTutorialMinimized(false);
  };

  return (
    <div className={styles.tutorialsMenu} ref={menuRef}>
      {tutorials.map((tutorial) => {
        return (
          <TutorialItem
            key={tutorial.id}
            content={tutorial.content}
            progress={tutorial.progress.initialStory}
            stories={tutorial.stories}
            tutorialComplete={
              tutorialProgress && tutorialProgress[tutorial.id]?.completed
            }
            tutorialStarted={tutorialProgress && tutorialProgress[tutorial.id]}
            handleClick={getHandleClick(tutorial.id)}
            handleDelete={() => {
              tutorial.cleanup.cleanupImageProcessing();
              deleteTutorialResources(tutorial.id);
              setTutorialsMenu(false);
            }}
            deleteLoading={tutorial.cleanup.loading}
            deleteError={tutorial.cleanup.error?.message}
            setStickTutorialsMenu={setStickTutorialsMenu}
          />
        );
      })}
    </div>
  );
};

export default TutorialsMenu;
