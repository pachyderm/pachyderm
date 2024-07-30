import escapeRegExp from 'lodash/escapeRegExp';
import React, {useCallback} from 'react';

import {
  ButtonLink,
  Icon,
  PipelineColorlessSVG,
  StatusWarningSVG,
} from '@pachyderm/components';

import styles from './SearchResultItem.module.css';

type SearchResultItemProps = {
  children?: React.ReactNode;
  title: string;
  searchValue: string;
  onClick: () => void;
  wasKilled?: boolean;
  hasFailedPipeline?: boolean;
  hasFailedSubjob?: boolean;
};

type SecondaryActionProps = {
  linkText?: string;
  onClick?: () => void;
};

const Underline = ({text = '', search = ''}) => {
  const regex = new RegExp(`^(${escapeRegExp(search)})`, 'gi');
  const parts = text.split(regex);
  return (
    <span className={styles.underlineContainer} title={text}>
      {parts.map((part, i) =>
        search && regex.test(part) ? (
          <span className={styles.underline} key={i}>
            {part}
          </span>
        ) : (
          <span key={i}>{part}</span>
        ),
      )}
    </span>
  );
};

export const SearchResultItem: React.FC<SearchResultItemProps> = ({
  title,
  searchValue,
  onClick,
  children,
  wasKilled,
  hasFailedPipeline,
  hasFailedSubjob,
}) => {
  return (
    <li
      className={styles.base}
      onClick={onClick}
      data-testid="SearchResultItem__container"
    >
      {title && <Underline text={title} search={searchValue} />}
      {children}
      {(hasFailedPipeline || hasFailedSubjob) && (
        <div className={styles.icons}>
          {wasKilled && <span className={styles.killed}>Killed</span>}
          {hasFailedPipeline && (
            <Icon color="red" small>
              <PipelineColorlessSVG />
            </Icon>
          )}
          {hasFailedSubjob && (
            <Icon color="red" small>
              <StatusWarningSVG />
            </Icon>
          )}
        </div>
      )}
    </li>
  );
};

export const SecondaryAction: React.FC<SecondaryActionProps> = ({
  linkText,
  onClick,
}) => {
  const onClickHandler = useCallback(
    (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
      if (onClick) {
        e.stopPropagation();
        onClick();
      }
    },
    [onClick],
  );

  return (
    <ButtonLink
      small
      onClick={onClickHandler}
      className={styles.secondaryAction}
    >
      {linkText}
    </ButtonLink>
  );
};
