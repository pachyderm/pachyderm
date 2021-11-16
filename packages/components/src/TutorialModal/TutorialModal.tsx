import classNames from 'classnames';
import React, {useMemo} from 'react';

import {Button} from 'Button';
import {ArrowRightSVG, ChevronUpSVG, ChevronDownSVG, MinimizeSVG} from 'Svg';
import {TaskCard} from 'TutorialModal';

import {PureCheckbox} from '../Checkbox';

import SideBar from './components/SideBar';
import TaskListItem from './components/TaskListItem';
import useTutorialModal from './hooks/useTutorialModal';
import {Step} from './lib/types';
import styles from './TutorialModal.module.css';

type TutorialModalProps = {
  steps: Step[];
  initialStep?: number;
  iniitalTask?: number;
};

const TutorialModal: React.FC<TutorialModalProps> = ({
  steps,
  initialStep = 0,
  iniitalTask = 0,
}) => {
  const {
    currentStep,
    currentTask,
    displayTaskIndex,
    displayTaskInstance,
    handleTaskCompletion,
    handleNextStep,
    minimized,
    nextTaskIndex,
    setMinimized,
  } = useTutorialModal(steps, initialStep, iniitalTask);

  const classes = classNames(styles.modal, {
    [styles.minimize]: minimized,
    [styles.maximize]: !minimized,
  });

  const nextTaskInstance = useMemo(() => {
    if (currentTask === steps[currentStep].sections.length) {
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
          task={steps[currentStep].sections[nextTaskIndex].taskName}
        />
      );
    }
  }, [nextTaskIndex, currentTask, currentStep, steps]);

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
            {nextTaskInstance}
          </div>
        ) : null}
        <div className={styles.header}>
          <Button
            className={styles.button}
            onClick={handleNextStep}
            disabled={
              currentStep === steps.length - 1 ||
              currentTask <= steps[currentStep]?.sections.length - 1
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
            currentStep={currentStep}
            currentTask={currentTask}
            steps={steps}
          />
          <div className={styles.content}>
            {steps[currentStep].sections.map((section, i) => {
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
              index={steps[currentStep].sections.length}
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

export default TutorialModal;
