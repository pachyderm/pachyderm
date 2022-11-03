import classNames from 'classnames';
import React, {useCallback} from 'react';

import styles from './Search.module.css';

interface SearchProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  onSearch: (val: string) => void;
}

export const Search: React.FC<SearchProps> = ({
  onSearch,
  className,
  ...props
}) => {
  const handleSearch = useCallback(
    (evt: React.ChangeEvent<HTMLInputElement>) => {
      onSearch(evt.target.value);
    },
    [onSearch],
  );
  return (
    <input
      role="searchbox"
      type="text"
      className={classNames(styles.base, className)}
      onChange={handleSearch}
      {...props}
    />
  );
};

export default Search;
