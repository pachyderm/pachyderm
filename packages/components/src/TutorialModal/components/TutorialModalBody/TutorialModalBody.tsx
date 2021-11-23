import classNames from 'classnames';
import React, {ReactNode} from 'react';

import {Button} from 'Button';
import {ArrowRightSVG, ChevronUpSVG, ChevronDownSVG, MinimizeSVG} from 'Svg';
import {TaskCard} from 'TutorialModal';

import {PureCheckbox} from '../../../Checkbox';
import SideBar from '../../components/SideBar';
import TaskListItem from '../../components/TaskListItem';
import {Story} from '../../lib/types';

import useTutorialModal from './hooks/useTutorialModal';
import styles from './TutorialModalBody.module.css';

type TutorialModalBodyProps = {
  stories: Story[];
  initialStory?: number;
  iniitalTask?: number;
};

const NextTaskInstance: React.FC<{
  currentTask: number;
  nextTaskIndex: number;
  storyLength: number;
  taskName?: ReactNode;
}> = ({currentTask, nextTaskIndex, storyLength, taskName}) => {
  if (currentTask === storyLength) {
    return (
      <PureCheckbox
        name="Continue"
        selected={false}
        label={'Continue to the next story'}
      />
    );
  } else if (currentTask === 0) {
    return null;
  } else {
    return (
      <TaskListItem
        currentTask={currentTask}
        index={nextTaskIndex}
        task={taskName}
      />
    );
  }
};

const TutorialModalBody: React.FC<TutorialModalBodyProps> = ({
  stories,
  initialStory = 0,
  iniitalTask = 0,
}) => {
  const {
    currentStory,
    currentTask,
    displayTaskIndex,
    displayTaskInstance,
    handleStoryChange,
    handleTaskCompletion,
    handleNextStep,
    minimized,
    nextTaskIndex,
    setMinimized,
  } = useTutorialModal(stories, initialStory, iniitalTask);

  const classes = classNames(styles.modal, {
    [styles.minimize]: minimized,
    [styles.maximize]: !minimized,
  });

  return (
    <>
      <div className={!minimized ? styles.overlay : ''} />
      <div className={classes}>
        {displayTaskInstance ? (
          <div
            className={classNames(styles.miniTask, {
              [styles.miniTaskOpen]: minimized,
            })}
          >
            <TaskListItem
              currentTask={currentTask}
              index={displayTaskIndex}
              task={displayTaskInstance}
            />
            <NextTaskInstance
              currentTask={currentTask}
              nextTaskIndex={nextTaskIndex}
              storyLength={stories[currentStory].sections.length}
              taskName={
                stories[currentStory]?.sections[nextTaskIndex]?.taskName
              }
            />
          </div>
        ) : null}
        <div className={styles.header}>
          <Button
            className={styles.button}
            onClick={handleNextStep}
            disabled={
              currentStory === stories.length - 1 ||
              currentTask <= stories[currentStory]?.sections.length - 1
            }
          >
            Next Story
            <ArrowRightSVG />
          </Button>
          <Button
            className={styles.button}
            buttonType="secondary"
            onClick={() => setMinimized((prevValue) => !prevValue)}
          >
            {minimized ? (
              <>
                Maximize
                <ChevronUpSVG />
              </>
            ) : (
              <>
                Minimize
                <ChevronDownSVG />
              </>
            )}
          </Button>
        </div>
        <div className={styles.body}>
          <SideBar
            currentStory={currentStory}
            currentTask={currentTask}
            stories={stories}
            handleStoryChange={handleStoryChange}
          />
          <div className={styles.content}>
            {stories[currentStory].sections.map((section, i) => {
              return (
                <div className={styles.section} key={i}>
                  <div className={styles.headerInfo}>
                    <div
                      className={
                        section.isSubHeader
                          ? styles.sectionSubHeader
                          : styles.sectionHeader
                      }
                    >
                      {section.header}
                    </div>
                    {section.info}
                  </div>
                  <section.Task
                    currentTask={currentTask}
                    onCompleted={() => handleTaskCompletion(i)}
                    minimized={minimized}
                    index={i}
                    name={section.taskName}
                  />
                  <div className={styles.followUp}>{section.followUp}</div>
                </div>
              );
            })}
            <TaskCard
              index={stories[currentStory].sections.length}
              currentTask={currentTask}
              task={
                <div>
                  <MinimizeSVG className={styles.minimizeSVG} /> Contine to the
                  next story
                </div>
              }
              action={handleNextStep}
              actionText={
                <>
                  Next Story
                  <ArrowRightSVG className={styles.rightArrow} />
                </>
              }
            />
          </div>
        </div>
      </div>
    </>
  );
};

export default TutorialModalBody;
