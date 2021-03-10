import classnames from 'classnames';
import React, {ButtonHTMLAttributes, useRef} from 'react';

import useDropdownButton from 'Dropdown/hooks/useDropdownButton';
import {ChevronDownSVG} from 'Svg';

import styles from './DropdownButton.module.css';

export interface DropdownButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement> {
  color?: 'purple' | 'black';
  hideChevron?: boolean;
}

export const DropdownButton: React.FC<DropdownButtonProps> = ({
  children,
  className,
  color = 'black',
  hideChevron = false,
  ...rest
}) => {
  const dropdownButtonRef = useRef<HTMLButtonElement>(null);
  const {toggleDropdown, isOpen, handleKeyDown} = useDropdownButton(
    dropdownButtonRef,
  );
  const mergedClasses = classnames(styles.base, className, {
    [styles.hideChevron]: hideChevron,
    [styles.purple]: color === 'purple',
  });

  return (
    <button
      ref={dropdownButtonRef}
      aria-haspopup
      aria-expanded={isOpen}
      className={mergedClasses}
      onClick={toggleDropdown}
      onKeyDown={handleKeyDown}
      type="button"
      {...rest}
    >
      <span className={styles.children}>{children}</span>

      {!hideChevron && <ChevronDownSVG aria-hidden className={styles.icon} />}
    </button>
  );
};

export default DropdownButton;
