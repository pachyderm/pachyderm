import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {useContext} from 'react';

import {Button} from 'Button';
import {Group} from 'Group';
import {Icon} from 'Icon';
import {StoryProgressDots} from 'StoryProgressDots';
import {
  ArrowRightSVG,
  ChevronUpSVG,
  ChevronDownSVG,
  MinimizeSVG,
  StatusPausedSVG,
} from 'Svg';
import {TaskCard} from 'TutorialModal';

import SideBar from '../../components/SideBar';
import {Story} from '../../lib/types';

import useTutorialModal from './hooks/useTutorialModal';
import styles from './TutorialModalBody.module.css';
import TutorialModalBodyContext from './TutorialModalBodyContext';

export type TutorialModalBodyProps = {
  stories: Story[];
  initialStory?: number;
  initialTask?: number;
  onTutorialComplete?: () => void;
  onSkip?: () => void;
  onClose?: () => void;
};

const TutorialModalBody: React.FC<TutorialModalBodyProps> = ({
  stories,
  initialStory = 0,
  initialTask = 0,
  onTutorialComplete = noop,
  onSkip,
  onClose,
}) => {
  const {
    currentStory,
    currentTask,
    handleNextStory,
    handleStoryChange,
    handleTaskCompletion,
    taskSections,
    tutorialModalRef,
  } = useTutorialModal(stories, initialStory, initialTask);

  const {minimized, setMinimized} = useContext(TutorialModalBodyContext);

  const classes = classNames(styles.modal, {
    [styles.minimize]: minimized,
    [styles.maximize]: !minimized,
  });

  return (
    <>
      <div className={!minimized ? styles.overlay : ''} />
      <div className={classes}>
        <div className={styles.header}>
          <Group spacing={16}>
            <div className={styles.storyProgressWrapper}>
              <StoryProgressDots
                dotStyle="light"
                stories={stories.length}
                progress={currentStory}
              />
            </div>
          </Group>
          <div className={styles.rightButtons}>
            {onSkip && (
              <Button
                className={styles.button}
                buttonType="secondary"
                onClick={onSkip}
                data-testid={'TutorialModalBody__skipTutorial'}
              >
                End Tutorial
              </Button>
            )}
            {onClose && (
              <Button
                className={styles.button}
                buttonType="secondary"
                onClick={onClose}
                data-testid={'TutorialModalBody__closeTutorial'}
              >
                <Icon small color="plum" className={styles.pauseIcon}>
                  <StatusPausedSVG />
                </Icon>
              </Button>
            )}
            <Button
              className={styles.button}
              buttonType="secondary"
              onClick={() => setMinimized(!minimized)}
              data-testid={`TutorialModalBody__${
                minimized ? 'maximize' : 'minimize'
              }`}
              aria-label={minimized ? 'maximize' : 'minimize'}
              IconSVG={minimized ? ChevronUpSVG : ChevronDownSVG}
            />
          </div>
        </div>
        <div className={styles.body} ref={tutorialModalRef}>
          <SideBar
            currentStory={currentStory}
            currentTask={currentTask}
            stories={stories}
            handleStoryChange={handleStoryChange}
            taskSections={taskSections}
          />
          <div className={styles.content}>
            {stories[currentStory].sections.map((section, i) => {
              return (
                <div className={styles.section} key={i}>
                  <div className={styles.headerInfo}>
                    {section.isSubHeader ? (
                      <h6 className={styles.sectionSubHeader}>
                        {section.header}
                      </h6>
                    ) : (
                      <h3 className={styles.sectionHeader}>{section.header}</h3>
                    )}
                    {section.info}
                  </div>
                  {section.Task && (
                    <section.Task
                      currentTask={currentTask}
                      currentStory={currentStory}
                      onCompleted={() =>
                        handleTaskCompletion(
                          taskSections.findIndex(
                            (taskSection) =>
                              taskSection.taskName === section.taskName,
                          ),
                        )
                      }
                      minimized={minimized}
                      index={taskSections.findIndex(
                        (taskSection) =>
                          taskSection.taskName === section.taskName,
                      )}
                      name={section.taskName}
                    />
                  )}
                  {section.followUp && (
                    <div className={styles.followUp}>{section.followUp}</div>
                  )}
                </div>
              );
            })}
            {currentStory < stories.length - 1 ? (
              <TaskCard
                index={taskSections.length}
                currentTask={currentTask}
                task={
                  <div>
                    <MinimizeSVG className={styles.minimizeSVG} /> Contine to
                    the next story
                  </div>
                }
                action={handleNextStory}
                actionText={
                  <>
                    Next Story
                    <ArrowRightSVG className={styles.rightArrow} />
                  </>
                }
              />
            ) : (
              <TaskCard
                index={taskSections.length}
                currentTask={currentTask}
                task={<div>Tutorial Complete!</div>}
                action={onTutorialComplete}
                actionText={<>Close Tutorial</>}
              />
            )}
          </div>
        </div>
      </div>
    </>
  );
};

export default TutorialModalBody;
