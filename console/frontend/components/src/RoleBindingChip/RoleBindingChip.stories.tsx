import React from 'react';

import {RoleBindingChip} from '.';

export default {title: 'RolesModal'};

export const Default = () => {
  return (
    <RoleBindingChip
      roleName="default"
      deleteRole={() => new Promise((res) => res)}
    />
  );
};

export const Locked = () => {
  return (
    <RoleBindingChip
      roleName="locked"
      locked
      deleteRole={() => new Promise((res) => res)}
    />
  );
};

export const ReadOnly = () => {
  return (
    <RoleBindingChip
      roleName="readOnly"
      readOnly
      deleteRole={() => new Promise((res) => res)}
    />
  );
};
