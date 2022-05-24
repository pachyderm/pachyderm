import classnames from 'classnames';
import React, {useRef, useMemo} from 'react';

import useDropdownButton from 'Dropdown/hooks/useDropdownButton';
import {ChevronDownSVG, ChevronRightSVG} from 'Svg';

import {Button, ButtonProps} from '../../../Button';

import styles from './DropdownButton.module.css';

export type DropdownButtonProps = Omit<
  ButtonProps,
  'href' | 'to' | 'download' | 'buttonRef'
> & {
  hideChevron?: boolean;
};

export const DropdownButton: React.FC<DropdownButtonProps> = ({
  children,
  className,
  hideChevron = false,
  disabled = false,
  buttonType = 'dropdown',
  IconSVG,
  ...rest
}) => {
  const dropdownButtonRef = useRef<HTMLButtonElement>(null);
  const {toggleDropdown, isOpen, handleKeyDown, sideOpen} = useDropdownButton(
    dropdownButtonRef,
  );
  const mergedClasses = classnames(styles.base, className);

  const Icon = useMemo<
    React.FunctionComponent<React.SVGProps<SVGSVGElement>> | undefined
  >(() => {
    if (IconSVG) {
      return IconSVG;
    }

    if (!hideChevron) {
      return sideOpen ? ChevronRightSVG : ChevronDownSVG;
    }

    return undefined;
  }, [IconSVG, hideChevron, sideOpen]);

  return (
    <Button
      data-testid="DropdownButton__button"
      buttonRef={dropdownButtonRef}
      aria-haspopup
      aria-expanded={isOpen}
      buttonType={buttonType}
      className={mergedClasses}
      onClick={toggleDropdown}
      onKeyDown={handleKeyDown}
      type="button"
      disabled={disabled}
      iconPosition="end"
      IconSVG={Icon}
      {...rest}
    >
      {children && <span className={styles.children}>{children}</span>}
    </Button>
  );
};

export default DropdownButton;
