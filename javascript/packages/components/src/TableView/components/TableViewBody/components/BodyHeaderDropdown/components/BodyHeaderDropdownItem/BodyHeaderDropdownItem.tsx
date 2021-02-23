import React from 'react';

import {Checkbox} from 'Checkbox';
import Dropdown from 'Dropdown';

import styles from './BodyHeaderDropdownItem.module.css';

export type BodyHeaderDropdownItemProps = {
  id: string;
  name: string;
  label: string;
};

const BodyHeaderDropdownItem: React.FC<BodyHeaderDropdownItemProps> = ({
  id,
  name,
  label,
}) => {
  return (
    <>
      <Dropdown.MenuItem id={id} className={styles.base}>
        <Checkbox
          small
          id={id}
          name={name}
          label={label}
          className={styles.checkbox}
          defaultChecked={true}
        />
      </Dropdown.MenuItem>
    </>
  );
};

export default BodyHeaderDropdownItem;
