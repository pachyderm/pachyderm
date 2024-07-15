import classnames from 'classnames';
import React, {useRef, useMemo} from 'react';

import {
  ChevronDownSVG,
  ChevronRightSVG,
  ChevronUpSVG,
} from '@pachyderm/components';
import useSideNav from '@pachyderm/components/SideNav/hooks/useSideNav';

import {Button, ButtonProps} from '../../../Button';
import useDropdownButton from '../../hooks/useDropdownButton';

import styles from './DropdownButton.module.css';

export type DropdownButtonProps = Omit<
  ButtonProps,
  'href' | 'to' | 'download' | 'buttonRef'
> & {
  hideChevron?: boolean;
  openOnClick?: () => void;
};

export const DropdownButton: React.FC<DropdownButtonProps> = ({
  children,
  className,
  hideChevron = false,
  disabled = false,
  buttonType = 'dropdown',
  IconSVG,
  openOnClick,
  ...rest
}) => {
  const dropdownButtonRef = useRef<HTMLButtonElement>(null);
  const {minimized} = useSideNav();
  const {toggleDropdown, isOpen, handleKeyDown, sideOpen, openUpwards} =
    useDropdownButton(dropdownButtonRef);
  const mergedClasses = classnames(styles.base, className);

  const Icon = useMemo<
    React.FunctionComponent<React.SVGProps<SVGSVGElement>> | undefined
  >(() => {
    if (IconSVG) {
      return IconSVG;
    }

    if (hideChevron) {
      return undefined;
    }

    if (sideOpen) {
      return ChevronRightSVG;
    }

    if (openUpwards) {
      return ChevronUpSVG;
    }

    return ChevronDownSVG;
  }, [IconSVG, hideChevron, sideOpen, openUpwards]);

  return (
    <Button
      data-testid="DropdownButton__button"
      buttonRef={dropdownButtonRef}
      aria-haspopup
      aria-expanded={isOpen}
      buttonType={buttonType}
      className={mergedClasses}
      onClick={(e) => {
        openOnClick && !isOpen && openOnClick();
        toggleDropdown(e);
      }}
      onKeyDown={handleKeyDown}
      type="button"
      disabled={disabled}
      iconPosition="end"
      IconSVG={Icon}
      {...rest}
    >
      {children && !minimized && (
        <span className={styles.children}>{children}</span>
      )}
    </Button>
  );
};

export default DropdownButton;
