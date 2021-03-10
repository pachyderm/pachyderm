import classnames from 'classnames';
import React from 'react';

import {useDropdown} from 'Dropdown';

import styles from './DropdownMenu.module.css';

export interface DropdownMenuProps {
  className?: string;
  pin?: 'left' | 'right';
  openAbove?: boolean;
}

export const DropdownMenu: React.FC<DropdownMenuProps> = ({
  children,
  className,
  pin = 'left',
  openAbove = false,
  ...rest
}) => {
  const {isOpen} = useDropdown();

  const classes = classnames(styles.base, className, {
    [styles.open]: isOpen,
    [styles.openAbove]: openAbove,
    [styles.left]: pin === 'left',
    [styles.right]: pin === 'right',
  });

  return (
    <div role="menu" className={classes} aria-hidden={!isOpen} {...rest}>
      {children}
    </div>
  );
};

export default DropdownMenu;
