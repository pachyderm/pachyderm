import capitalize from 'lodash/capitalize';
import React from 'react';

import {ResourceType} from '@dash-frontend/api/auth';
import BrandedDocLink from '@dash-frontend/components/BrandedDocLink';
import {REPO_ROLES, ALL_ROLES} from '@dash-frontend/constants/rbac';
import {
  Icon,
  InfoSVG,
  TrashSVG,
  BasicModal,
  AddCircleSVG,
  Group,
  HelpText,
  Tooltip,
  CaptionTextSmall,
  DefaultDropdown,
  RoleBindingChip,
  SearchSVG,
  Button,
  CloseSVG,
  UndoSVG,
  LockSVG,
} from '@pachyderm/components';

import {AssignRolesForm} from './components/AssignRolesForm';
import useRolesModal from './hooks/useRolesModal';
import styles from './RolesModal.module.css';

type RolesModalProps = {
  resourceType: ResourceType.PROJECT | ResourceType.REPO;
  projectName: string;
  repoName?: string;
  disclaimerText?: string;
  show: boolean;
  onHide?: () => void;
  readOnly?: boolean;
  readOnlyText?: string;
};

export const RolesModal: React.FC<RolesModalProps> = ({
  resourceType,
  show,
  onHide,
  disclaimerText,
  projectName,
  repoName,
  readOnly,
  readOnlyText,
}) => {
  const {
    deletedRoles,
    setDeletedRoles,
    principalFilterRef,
    principalFilterOpen,
    openPrincipalFilter,
    exitPrincipalFilter,
    setPrincipalFilter,
    userTableRoles,
    loading,
    error,
    resourceName,
    deleteAllRoles,
    undoDeleteAllRoles,
    deleteRole,
    addSelectedRole,
    modifyRolesError,
  } = useRolesModal({projectName, repoName, resourceType});

  const mainResourceRoles =
    resourceType === ResourceType.REPO ? REPO_ROLES : ALL_ROLES;

  return (
    <BasicModal
      className={styles.base}
      show={show}
      onHide={onHide}
      headerContent={`${!readOnly ? 'Set ' : ''}${capitalize(
        resourceType,
      )} Level Roles: ${resourceName}`}
      loading={loading}
      errorMessage={error || modifyRolesError}
      actionable
      hideConfirm
      mode="Long"
      cancelText="Done"
      footerContent={
        <BrandedDocLink pathWithoutDomain="/set-up/authorization/">
          Info
        </BrandedDocLink>
      }
      flexBody
    >
      {!readOnly && (
        <AssignRolesForm
          roles={mainResourceRoles}
          resourceType={resourceType}
          resourceName={resourceName}
          userTableRoles={userTableRoles}
          deletedRoles={deletedRoles}
          setDeletedRoles={setDeletedRoles}
        />
      )}
      {disclaimerText && (
        <Group className={styles.disclaimerText} spacing={8}>
          <Icon small color="grey">
            <InfoSVG />
          </Icon>
          <HelpText>{disclaimerText}</HelpText>
        </Group>
      )}
      {readOnly && readOnlyText && (
        <Group className={styles.disclaimerText} spacing={8}>
          <Icon small color="grey">
            <LockSVG />
          </Icon>
          <HelpText>{readOnlyText}</HelpText>
        </Group>
      )}
      <table className={styles.flexTable}>
        <thead className={styles.headerRow}>
          <tr>
            {principalFilterOpen && (
              <th className={styles.columnHeader}>
                <Icon small>
                  <SearchSVG aria-hidden />
                </Icon>

                <input
                  ref={principalFilterRef}
                  className={styles.filterInput}
                  placeholder="Search Users"
                  role="searchbox"
                  onChange={(e) => setPrincipalFilter(e.target.value)}
                />

                <Button
                  aria-label="close users search"
                  buttonType="ghost"
                  onClick={exitPrincipalFilter}
                  IconSVG={CloseSVG}
                />
              </th>
            )}
            {!principalFilterOpen && (
              <th className={styles.columnHeader}>
                <span>Name</span>
                <Button
                  aria-label="open users search"
                  buttonType="ghost"
                  onClick={openPrincipalFilter}
                  IconSVG={SearchSVG}
                />
              </th>
            )}
            <th className={styles.columnHeader}>
              {capitalize(resourceType)} Roles
            </th>
          </tr>
        </thead>
        <tbody className={styles.rolesTable}>
          {Object.keys(userTableRoles)?.map((principal) => {
            const availableRoles = mainResourceRoles
              .filter(
                (role) =>
                  !userTableRoles[principal].unlockedRoles.includes(role) &&
                  !userTableRoles[principal].lockedRoles.includes(role),
              )
              .map((role) => ({
                id: role,
                content: role,
                closeOnClick: true,
              }));

            return (
              <tr className={styles.userRow} key={principal}>
                <td>
                  <div className={styles.userCell}>
                    {!readOnly &&
                      !deletedRoles[principal] &&
                      userTableRoles[principal].unlockedRoles.length > 0 && (
                        <Tooltip
                          tooltipText={`Delete all
                          ${resourceType.toLowerCase()} roles`}
                        >
                          <Icon
                            small
                            color="red"
                            className={styles.cellIcon}
                            onClick={deleteAllRoles(
                              principal,
                              userTableRoles[principal].unlockedRoles,
                            )}
                            aria-label="delete all roles"
                            role="button"
                          >
                            <TrashSVG />
                          </Icon>
                        </Tooltip>
                      )}
                    {!readOnly && deletedRoles[principal] && (
                      <Tooltip tooltipText="Undo Delete">
                        <Icon
                          small
                          color="black"
                          className={styles.cellIcon}
                          onClick={undoDeleteAllRoles(principal)}
                          aria-label="undo delete all roles"
                          role="button"
                        >
                          <UndoSVG />
                        </Icon>
                      </Tooltip>
                    )}
                    {principal}
                  </div>
                  {userTableRoles[principal].higherRoles.length > 0 && (
                    <div>
                      {userTableRoles[principal].higherRoles[0] && (
                        <CaptionTextSmall>
                          {userTableRoles[principal].higherRoles[0]}
                        </CaptionTextSmall>
                      )}
                      {userTableRoles[principal].higherRoles[1] && (
                        <CaptionTextSmall>
                          , {userTableRoles[principal].higherRoles[1]}
                        </CaptionTextSmall>
                      )}
                      {userTableRoles[principal].higherRoles.length > 2 && (
                        <Tooltip
                          tooltipText={userTableRoles[principal].higherRoles
                            .slice(2)
                            .join(', ')}
                        >
                          <CaptionTextSmall>
                            , +
                            {userTableRoles[principal].higherRoles.length - 2}{' '}
                            more
                          </CaptionTextSmall>
                        </Tooltip>
                      )}
                    </div>
                  )}
                </td>
                <td>
                  {userTableRoles[principal].lockedRoles.map(
                    (role: string | undefined) =>
                      role ? (
                        <RoleBindingChip key={role} roleName={role} locked />
                      ) : null,
                  )}
                  {userTableRoles[principal].unlockedRoles.map(
                    (role: string | undefined) =>
                      role ? (
                        <RoleBindingChip
                          key={role}
                          roleName={role}
                          readOnly={readOnly}
                          deleteRole={deleteRole(
                            principal,
                            userTableRoles[principal].unlockedRoles,
                            role,
                          )}
                        />
                      ) : null,
                  )}
                  {!readOnly && availableRoles.length > 0 && (
                    <DefaultDropdown
                      className={styles.typeDropdown}
                      items={availableRoles}
                      aria-label={`add role to ${principal}`}
                      onSelect={addSelectedRole(
                        principal,
                        userTableRoles[principal].unlockedRoles,
                      )}
                      buttonOpts={{
                        className: styles.addCircle,
                        hideChevron: true,
                        IconSVG: AddCircleSVG,
                        buttonType: 'ghost',
                      }}
                      menuOpts={{pin: 'right'}}
                    />
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </BasicModal>
  );
};

export default RolesModal;
