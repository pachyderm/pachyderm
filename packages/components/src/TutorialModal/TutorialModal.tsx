import classNames from 'classnames';
import React, {useMemo} from 'react';

import {Button} from 'Button';
import {ArrowRightSVG, ChevronUpSVG, ChevronDownSVG} from 'Svg';

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
    if (currentTask === steps[currentStep].tasks.length) {
      return (
        <PureCheckbox
          name="Continue"
          selected={false}
          label={'Continue to next step.'}
        />
      );
    } else if (currentTask === 0) {
      return null;
    } else {
      return (
        <TaskListItem
          currentTask={currentTask}
          index={nextTaskIndex}
          task={steps[currentStep].tasks[nextTaskIndex]}
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
              currentTask <= steps[currentStep]?.tasks.length - 1
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
            <div className={styles.instructions}>
              <h2 className={styles.instructionsHeader}>
                {steps[currentStep].instructionsHeader}
              </h2>
              <div className={styles.instructionsBody}>
                <div className={styles.instructionsText}>
                  {steps[currentStep].instructionsText}
                </div>
              </div>
            </div>
            {steps[currentStep].tasks.map((task, i) => {
              return (
                <React.Fragment key={task.name?.toString()}>
                  <task.Task
                    currentTask={currentTask}
                    onCompleted={() => handleTaskCompletion(i)}
                    minimized={minimized}
                    index={i}
                    name={task.name}
                  />
                  <div className={styles.followUp}>{task.followUp}</div>
                </React.Fragment>
              );
            })}
          </div>
        </div>
      </div>
    </>
  );
};

export default TutorialModal;
