import classnames from 'classnames';
import noop from 'lodash/noop';
import React, {ButtonHTMLAttributes, useRef} from 'react';

import useDropdownMenuItem from 'Dropdown/hooks/useDropdownMenuItem';

import styles from './DropdownMenuItem.module.css';

export interface DropdownMenuItemProps
  extends ButtonHTMLAttributes<HTMLButtonElement> {
  id: string;
  important?: boolean;
  closeOnClick?: boolean;
}

export const DropdownMenuItem: React.FC<DropdownMenuItemProps> = ({
  children,
  important = false,
  id,
  className,
  onClick = noop,
  closeOnClick = false,
  ...rest
}) => {
  const ref = useRef<HTMLButtonElement>(null);
  const {isSelected, handleClick, handleKeyDown} = useDropdownMenuItem({
    id,
    onClick,
    closeOnClick,
    ref,
  });
  const classes = classnames(styles.base, className, {
    [styles.important]: important,
    [styles.selected]: isSelected,
  });

  return (
    <button
      ref={ref}
      data-testid="DropdownMenuItem__button"
      role="menuitem"
      type="button"
      className={classes}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      tabIndex={-1}
      {...rest}
    >
      {children}
    </button>
  );
};

export default DropdownMenuItem;
