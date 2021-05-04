import classnames from 'classnames';
import React, {HTMLAttributes, useMemo} from 'react';
import {useFormContext} from 'react-hook-form';

import useFilteredResults from 'Dropdown/hooks/useFilteredResults';

import parentStyles from '../DropdownMenuItem/DropdownMenuItem.module.css';

import styles from './DropdownMenuEmptyResults.module.css';

export type DropdownMenuEmptyResultsProps = HTMLAttributes<HTMLDivElement>;

const DropdownMenuEmptyResults: React.FC<HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
}) => {
  const {watch} = useFormContext();
  const {filteredResults} = useFilteredResults();

  const searchValue = watch('search');

  const shown = useMemo(() => searchValue && filteredResults.length === 0, [
    searchValue,
    filteredResults,
  ]);

  if (!shown) {
    return null;
  }

  return (
    <div className={classnames(parentStyles.base, styles.base, className)}>
      {children}
    </div>
  );
};

export default DropdownMenuEmptyResults;
