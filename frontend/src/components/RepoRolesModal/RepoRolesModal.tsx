import React, {FunctionComponent, useEffect} from 'react';

import {ResourceType} from '@dash-frontend/api/auth';

import {RolesModal} from '../RolesModal';

type RepoRolesModalProps = {
  show: boolean;
  onHide?: () => void;
  projectName: string;
  repoName: string;
  readOnly?: boolean;
  checkPermissions?: () => void;
};

const RepoRolesModal: FunctionComponent<RepoRolesModalProps> = ({
  show,
  onHide,
  projectName,
  repoName,
  readOnly,
  checkPermissions,
}) => {
  const disclaimerText =
    'Pipelines inherit the roles of the output repo, which means users will have the same level of access to both.';
  const readOnlyText = 'You need at least repoOwner to edit roles';

  useEffect(() => {
    checkPermissions && checkPermissions();
  }, [checkPermissions]);

  return (
    <RolesModal
      resourceType={ResourceType.REPO}
      projectName={projectName}
      repoName={repoName}
      disclaimerText={disclaimerText}
      show={show}
      onHide={onHide}
      readOnly={readOnly}
      readOnlyText={readOnlyText}
    />
  );
};

export default RepoRolesModal;
