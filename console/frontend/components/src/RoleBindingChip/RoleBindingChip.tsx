import React from 'react';

import {Icon, CloseSVG, LockSVG} from '@pachyderm/components';

import styles from './RoleBindingChip.module.css';

type RoleBindingChipProps = {
  roleName: string;
  readOnly?: boolean;
  locked?: boolean;
  deleteRole?: () => Promise<void>;
};

export const RoleBindingChip: React.FC<RoleBindingChipProps> = ({
  roleName,
  readOnly,
  locked,
  deleteRole,
}) => {
  const getIcon = () => {
    if (readOnly) return null;
    if (locked) {
      return (
        <Icon
          small
          color="grey"
          className={styles.locked}
          aria-label={`role ${roleName} locked`}
        >
          <LockSVG />
        </Icon>
      );
    } else {
      return (
        <Icon
          small
          color="black"
          onClick={deleteRole}
          className={styles.onDelete}
          aria-label={`delete role ${roleName}`}
        >
          <CloseSVG />
        </Icon>
      );
    }
  };

  return (
    <div className={styles.base}>
      {roleName}
      {getIcon()}
    </div>
  );
};

export default RoleBindingChip;
