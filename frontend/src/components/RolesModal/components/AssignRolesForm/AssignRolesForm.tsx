import classNames from 'classnames';
import React from 'react';

import {ModifyRoleBindingRequest, ResourceType} from '@dash-frontend/api/auth';
import {
  Label,
  DefaultDropdown,
  Button,
  Group,
  CaptionTextSmall,
  SpinnerSVG,
  Icon,
  StatusWarningSVG,
  ErrorText,
} from '@pachyderm/components';

import {UserTableRoles} from '../../hooks/useRolesModal';

import styles from './AssignRolesForm.module.css';
import useAssignRolesForm from './hooks/useAssignRolesForm';

type AssignRolesFormProps = {
  roles: string[];
  resourceType: ResourceType.PROJECT | ResourceType.REPO;
  resourceName: string;
  userTableRoles: UserTableRoles;
  deletedRoles: Record<string, ModifyRoleBindingRequest>;
  setDeletedRoles: React.Dispatch<
    React.SetStateAction<Record<string, ModifyRoleBindingRequest>>
  >;
};

export const USER_TYPES = ['user', 'group', 'robot'];

export const AssignRolesForm: React.FC<AssignRolesFormProps> = ({
  roles,
  resourceType,
  resourceName,
  userTableRoles,
  deletedRoles,
  setDeletedRoles,
}) => {
  const {
    inputRef,
    email,
    setEmail,
    role,
    setRole,
    userType,
    setUserType,
    validationError,
    error,
    loading,
    onSubmit,
    hasAllClusterUsers,
  } = useAssignRolesForm(
    resourceName,
    resourceType,
    roles,
    userTableRoles,
    deletedRoles,
    setDeletedRoles,
  );

  const userTypeOptions = [...USER_TYPES];
  if (!hasAllClusterUsers) {
    userTypeOptions.push('allClusterUsers');
  }

  return (
    <>
      <Label label="Assign roles" htmlFor="email" />
      <Group spacing={8} className={styles.base}>
        <span className={styles.mainInput}>
          <DefaultDropdown
            className={classNames(styles.typeDropdown, {
              [styles.allClusterUsers]: userType === 'allClusterUsers',
            })}
            items={userTypeOptions.map((type) => ({
              id: type,
              content: type,
              closeOnClick: true,
            }))}
            onSelect={(type) => setUserType(type)}
            buttonOpts={{
              buttonType: 'dropdown',
            }}
            menuOpts={{pin: 'left'}}
          >
            {userType}
          </DefaultDropdown>
          {userType !== 'allClusterUsers' && (
            <input
              ref={inputRef}
              autoComplete="off"
              placeholder="Add a single name or email address"
              className={styles.emailInput}
              onChange={(e) => setEmail(e.target.value)}
            />
          )}
        </span>
        <DefaultDropdown
          items={roles.map((role) => ({
            id: role,
            content: role,
            closeOnClick: true,
          }))}
          onSelect={(role) => setRole(role)}
          buttonOpts={{
            buttonType: 'dropdown',
          }}
          menuOpts={{pin: 'left'}}
        >
          {role}
        </DefaultDropdown>
        <Button type="submit" onClick={onSubmit}>
          Add
        </Button>
      </Group>
      {loading && (
        <Group spacing={8} className={styles.status}>
          <Icon small>
            <SpinnerSVG />
          </Icon>
          <CaptionTextSmall color="black">
            Adding {email || 'user'} to {resourceType.toLowerCase()}
          </CaptionTextSmall>
        </Group>
      )}
      {validationError && (
        <Group spacing={8} className={styles.status}>
          <Icon small color="red">
            <StatusWarningSVG />
          </Icon>
          <ErrorText color="black" role="alert" aria-label={validationError}>
            {validationError}
          </ErrorText>
        </Group>
      )}
      {error && (
        <Group spacing={8} className={styles.status}>
          <Icon small color="red">
            <StatusWarningSVG />
          </Icon>
          <ErrorText color="black" role="alert">
            Error adding {email || 'user'} to {resourceType.toLowerCase()}
          </ErrorText>
        </Group>
      )}
    </>
  );
};

export default AssignRolesForm;
