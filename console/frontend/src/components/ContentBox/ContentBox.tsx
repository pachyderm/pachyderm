import classNames from 'classnames';
import React from 'react';

import {CaptionTextSmall, Group, SkeletonBodyText} from '@pachyderm/components';

import styles from './ContentBox.module.css';

export type ContentBoxContent = {
  label: string;
  value: React.ReactNode;
  dataTestId?: string;
  responsiveValue?: React.ReactNode;
};

type ContentBoxProps = {
  content: ContentBoxContent[];
  loading?: boolean;
};

const ContentBox: React.FC<ContentBoxProps> = ({content, loading}) => {
  return (
    <Group className={styles.borderBox} spacing={4} vertical>
      {content.map(({label, value, dataTestId, responsiveValue}) => (
        <div
          key={label}
          className={styles.row}
          data-testid={dataTestId ? `${dataTestId}` : undefined}
        >
          {loading ? (
            <div className={styles.loading}>
              <SkeletonBodyText />
            </div>
          ) : (
            <>
              <CaptionTextSmall
                className={classNames(styles.label, {
                  [styles.full]: content.length === 1,
                })}
              >
                {label}:
              </CaptionTextSmall>
              <span
                className={classNames(styles.content, {
                  [styles.fullContent]: responsiveValue,
                  [styles.full]: content.length === 1,
                })}
              >
                {value || '-'}
              </span>
              {responsiveValue && (
                <span
                  className={classNames(
                    styles.content,
                    styles.responsiveContent,
                  )}
                >
                  {responsiveValue}
                </span>
              )}
            </>
          )}
        </div>
      ))}
    </Group>
  );
};

export default ContentBox;
