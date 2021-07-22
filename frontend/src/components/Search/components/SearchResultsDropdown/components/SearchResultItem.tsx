import {ButtonLink} from '@pachyderm/components';
import escapeRegExp from 'lodash/escapeRegExp';
import React, {useCallback} from 'react';

import styles from './SearchResultItem.module.css';

type SearchResultItemProps = {
  title: string;
  searchValue: string;
  onClick: () => void;
};

type SecondaryActionProps = {
  linkText?: string;
  onClick?: () => void;
};

const Underline = ({text = '', search = ''}) => {
  const regex = new RegExp(`^(${escapeRegExp(search)})`, 'gi');
  const parts = text.split(regex);
  return (
    <span>
      {parts.map((part, i) =>
        regex.test(part) ? (
          <div className={styles.underline} key={i}>
            {part}
          </div>
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
}) => {
  return (
    <div className={styles.base} onClick={onClick}>
      <Underline text={title} search={searchValue} />
      {children}
    </div>
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
