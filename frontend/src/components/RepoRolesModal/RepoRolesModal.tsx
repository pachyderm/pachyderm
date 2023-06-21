import {ResourceType} from '@graphqlTypes';
import React, {FunctionComponent} from 'react';

import {RolesModal} from '../RolesModal';

type RepoRolesModalProps = {
  show: boolean;
  onHide?: () => void;
  projectName: string;
  repoName: string;
  readOnly?: boolean;
};

const RepoRolesModal: FunctionComponent<RepoRolesModalProps> = ({
  show,
  onHide,
  projectName,
  repoName,
  readOnly,
}) => {
  const disclaimerText =
    'Pipelines inherit the roles of the output repo, which means users will have the same level of access to both.';
  const readOnlyText = 'You need at least repoOwner to edit roles';

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
