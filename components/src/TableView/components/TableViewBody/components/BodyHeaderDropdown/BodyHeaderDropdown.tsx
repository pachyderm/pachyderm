import React from 'react';
import {UseFormReturn} from 'react-hook-form';

import {Dropdown} from 'Dropdown';

import {DropdownButtonProps} from '../../../../../Dropdown/components/DropdownButton/DropdownButton';

import styles from './BodyHeaderDropdown.module.css';

export type BodyHeaderDropdownProps = {
  formCtx: UseFormReturn;
  buttonText: string;
  color?: DropdownButtonProps['color'];
  disabled?: DropdownButtonProps['disabled'];
};

const BodyHeaderDropdown: React.FC<BodyHeaderDropdownProps> = ({
  color,
  formCtx,
  buttonText,
  children,
  disabled,
}) => {
  return (
    <Dropdown formCtx={formCtx}>
      <Dropdown.Button
        className={styles.dropdownButton}
        color={color}
        disabled={disabled}
      >
        {buttonText}
      </Dropdown.Button>
      <Dropdown.Menu pin="right" className={styles.dropdownMenu}>
        {children}
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default BodyHeaderDropdown;
