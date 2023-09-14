import React, {useCallback} from 'react';

import {
  Button,
  CaptionTextSmall,
  CheckmarkSVG,
  TrashSVG,
  BasicModal,
  useModal,
  Icon,
  StoryProgressDots,
  Story,
} from '@pachyderm/components';

import styles from './TutorialItem.module.css';

const WARNING_MESSAGE =
  'All repositories, pipelines, and jobs that were part of this tutorial will be deleted permanently. You can restart tutorials from the "Run Tutorials" button.';

type TutorialItemProps = {
  content: string;
  stories: Story[];
  progress?: number;
  tutorialComplete: boolean;
  tutorialStarted: boolean;
  handleClick?: () => void;
  handleDelete?: () => void;
  deleteLoading?: boolean;
  deleteError?: string;
  setStickTutorialsMenu: React.Dispatch<React.SetStateAction<boolean>>;
};

const TutorialItem: React.FC<TutorialItemProps> = ({
  content,
  stories,
  progress,
  handleClick,
  handleDelete,
  deleteLoading,
  deleteError,
  setStickTutorialsMenu,
  tutorialComplete,
  tutorialStarted,
}) => {
  const {
    openModal: openConfirmationModal,
    isOpen: isConfirmationModalOpen,
    closeModal: closeConfirmationModal,
  } = useModal(false);

  const deleteHandler = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      setStickTutorialsMenu(true);
      openConfirmationModal();
    },
    [openConfirmationModal, setStickTutorialsMenu],
  );

  const clickHandler = useCallback(
    (e: React.MouseEvent) => {
      if (tutorialComplete) {
        deleteHandler(e);
      } else {
        handleClick && handleClick();
      }
    },
    [deleteHandler, handleClick, tutorialComplete],
  );

  return (
    <>
      <div role="button" className={styles.base} onClick={clickHandler}>
        <div className={styles.content}>
          {content}
          <div className={styles.progress}>
            {!tutorialStarted && (
              <CaptionTextSmall className={styles.subtitle}>
                Start Tutorial
              </CaptionTextSmall>
            )}
            {tutorialStarted && !tutorialComplete && (
              <StoryProgressDots progress={progress} stories={stories.length} />
            )}
            {tutorialComplete && (
              <CaptionTextSmall className={styles.subtitle}>
                <Icon color="highlightGreen">
                  <CheckmarkSVG />
                </Icon>
                Tutorial Completed{' '}
              </CaptionTextSmall>
            )}
          </div>
        </div>
        {progress !== undefined &&
          progress !== null &&
          !tutorialComplete &&
          handleDelete && (
            <Button
              data-testid="TutorialItem__deleteProgress"
              IconSVG={TrashSVG}
              buttonType="tertiary"
              onClick={deleteHandler}
            />
          )}
        {tutorialComplete && (
          <Icon color="white" small>
            <TrashSVG />
          </Icon>
        )}
      </div>

      {handleDelete && (
        <BasicModal
          show={isConfirmationModalOpen}
          headerContent={
            'You are about to delete all of your tutorial content.'
          }
          actionable
          loading={deleteLoading}
          errorMessage={deleteError}
          mode="Small"
          onHide={() => {
            closeConfirmationModal();
            setStickTutorialsMenu(false);
          }}
          onConfirm={() => {
            handleDelete();
            closeConfirmationModal();
            setStickTutorialsMenu(false);
          }}
          confirmText="Delete Tutorial Content"
        >
          {WARNING_MESSAGE}
        </BasicModal>
      )}
    </>
  );
};

export default TutorialItem;
