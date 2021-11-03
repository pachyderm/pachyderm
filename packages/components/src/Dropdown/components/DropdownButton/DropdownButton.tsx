import classnames from 'classnames';
import React, {ButtonHTMLAttributes, useRef} from 'react';

import useDropdownButton from 'Dropdown/hooks/useDropdownButton';
import {ChevronDownSVG, ChevronRightSVG} from 'Svg';

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
  disabled = false,
  ...rest
}) => {
  const dropdownButtonRef = useRef<HTMLButtonElement>(null);
  const {toggleDropdown, isOpen, handleKeyDown, sideOpen} = useDropdownButton(
    dropdownButtonRef,
  );
  const mergedClasses = classnames(styles.base, className, {
    [styles.hideChevron]: hideChevron,
    [styles.purple]: color === 'purple',
    [styles.disabled]: disabled,
    [styles.sideOpen]: sideOpen,
  });

  return (
    <button
      data-testid="DropdownButton__button"
      ref={dropdownButtonRef}
      aria-haspopup
      aria-expanded={isOpen}
      className={mergedClasses}
      onClick={toggleDropdown}
      onKeyDown={handleKeyDown}
      type="button"
      disabled={disabled}
      {...rest}
    >
      <span className={styles.children}>{children}</span>

      {!hideChevron &&
        (sideOpen ? (
          <ChevronRightSVG aria-hidden className={styles.icon} />
        ) : (
          <ChevronDownSVG aria-hidden className={styles.icon} />
        ))}
    </button>
  );
};

export default DropdownButton;
