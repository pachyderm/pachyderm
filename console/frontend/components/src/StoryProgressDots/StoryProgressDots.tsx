import classnames from 'classnames';
import times from 'lodash/times';
import React from 'react';

import {CaptionTextSmall} from '@pachyderm/components';

import styles from './StoryProgressDots.module.css';

type StoryProgressProps = {
  stories: number;
  progress?: number;
  dotStyle?: 'light' | 'dark';
};

const StoryProgressDots: React.FC<StoryProgressProps> = ({
  stories,
  progress = 0,
  dotStyle = 'dark',
}) => {
  return (
    <span
      className={classnames({
        [styles[dotStyle]]: true,
      })}
      data-testid="StoryProgressDots__progress"
    >
      {times(stories, (index) => {
        return (
          <span key={`${index}-dot`}>
            <span
              className={classnames(styles.progressDot, {
                [styles.progressMade]: progress >= index,
              })}
            />
            {index !== stories - 1 && (
              <span
                className={classnames(styles.progressConnector, {
                  [styles.progressMade]: progress >= index + 1,
                })}
              />
            )}
          </span>
        );
      })}
      <CaptionTextSmall className={styles.text}>{`Story ${
        progress + 1
      } of ${stories}`}</CaptionTextSmall>
    </span>
  );
};

export default StoryProgressDots;
