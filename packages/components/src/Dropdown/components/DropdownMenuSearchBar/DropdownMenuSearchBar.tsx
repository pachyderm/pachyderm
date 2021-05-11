import classnames from 'classnames';
import React, {InputHTMLAttributes, useCallback, useRef} from 'react';
import {useFormContext} from 'react-hook-form';

import useItemKeyController from 'Dropdown/hooks/useItemKeyController';

import {CloseSVG, SearchSVG} from '../../../Svg';

import styles from './DropdownMenuSearchBar.module.css';

export type DropdownMenuSearchBarProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  'name'
>;

const DropdownMenuSearchBar: React.FC<DropdownMenuSearchBarProps> = ({
  className,
  autoComplete = 'off',
  ...rest
}) => {
  const {register, setValue} = useFormContext();
  const ref = useRef<HTMLDivElement>(null);
  const {handleKeyDown} = useItemKeyController({ref});

  const clear = useCallback(() => {
    setValue('search', '');
  }, [setValue]);

  return (
    <div className={styles.base} ref={ref}>
      <SearchSVG className={styles.search} aria-hidden />

      <input
        className={classnames(styles.input, className)}
        id="search"
        name="search"
        aria-label="Search"
        ref={register}
        autoComplete={autoComplete}
        onKeyDown={handleKeyDown}
        {...rest}
      />

      <button aria-label="Clear" className={styles.close} onClick={clear}>
        <CloseSVG aria-hidden className={styles.closeIcon} />
      </button>
    </div>
  );
};

export default DropdownMenuSearchBar;
