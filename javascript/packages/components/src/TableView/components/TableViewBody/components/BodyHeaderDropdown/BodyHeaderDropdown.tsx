import React from 'react';
import {UseFormMethods} from 'react-hook-form';

import Dropdown from 'Dropdown';
import {DropdownButtonProps} from 'Dropdown/components/DropdownButton/DropdownButton';
import {Form} from 'Form';

import styles from './BodyHeaderDropdown.module.css';

export type BodyHeaderDropdownProps = {
  formCtx: UseFormMethods;
  buttonText: string;
  color?: DropdownButtonProps['color'];
};

const BodyHeaderDropdown: React.FC<BodyHeaderDropdownProps> = ({
  color,
  formCtx,
  buttonText,
  children,
}) => {
  return (
    <Form formContext={formCtx}>
      <Dropdown>
        <Dropdown.Button className={styles.dropdownButton} color={color}>
          {buttonText}
        </Dropdown.Button>
        <Dropdown.Menu pin="right" className={styles.dropdownMenu}>
          {children}
        </Dropdown.Menu>
      </Dropdown>
    </Form>
  );
};

export default BodyHeaderDropdown;
