import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {useMemo} from 'react';

import {StatusCheckmarkSVG} from '@pachyderm/components';

import {Circle} from '../../../Circle';
import {Icon} from '../../../Icon';
import useProgressBar from '../../hooks/useProgressBar';

import styles from './ProgressBarStep.module.css';

type Props = {
  children?: React.ReactNode;
  id: string;
  nextStepID?: string;
  onClick?: () => void;
};

const ProgressBarStep: React.FC<Props> = ({
  id,
  onClick,
  children,
  nextStepID,
}) => {
  const {isVisited, isCompleted, isVertical} = useProgressBar();
  const classes = useMemo(() => {
    return classNames(styles.base, {
      [styles.vertical]: isVertical,
      [styles.visited]: isVisited(id),
      [styles.nextStepVisited]: nextStepID && isVisited(nextStepID),
    });
  }, [id, isVisited, isVertical, nextStepID]);

  return (
    <div className={classes} data-testid="ProgressBarStep__div">
      <button
        data-testid="ProgressBarStep__group"
        onClick={isVisited(id) ? onClick : noop}
        className={styles.button}
        aria-disabled={!isVisited(id)}
        tabIndex={isVisited(id) ? 0 : -1}
      >
        <div className={styles.iconWrapper}>
          {isCompleted(id) ? (
            <Icon color="green" className={styles.checkmark} small>
              <StatusCheckmarkSVG
                aria-label="check mark"
                data-testid="ProgressBarStep__successCheckmark"
              />
            </Icon>
          ) : (
            <Circle className={styles.circle} />
          )}
        </div>

        {children}
      </button>
    </div>
  );
};

export default ProgressBarStep;
