import {ButtonLink} from '@pachyderm/components';
import escapeRegExp from 'lodash/escapeRegExp';
import React from 'react';

import styles from './SearchResultItem.module.css';

type SearchResultItemProps = {
  title: string;
  searchValue: string;
  linkText: string;
  onClick: () => void;
};

const SearchResultItem: React.FC<SearchResultItemProps> = ({
  title,
  linkText,
  searchValue,
  onClick,
}) => {
  return (
    <div className={styles.base}>
      <Underline text={title} search={searchValue} />
      <ButtonLink onClick={onClick} small>
        {linkText}
      </ButtonLink>
    </div>
  );
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

export default SearchResultItem;
