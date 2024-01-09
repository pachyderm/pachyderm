import React, {FunctionComponent} from 'react';

import {ResourceType} from '@dash-frontend/api/auth';

import {RolesModal} from '../RolesModal';

type ProjectRolesModalProps = {
  show: boolean;
  onHide?: () => void;
  projectName: string;
  readOnly?: boolean;
};

const ProjectRolesModal: FunctionComponent<ProjectRolesModalProps> = ({
  show,
  onHide,
  projectName,
  readOnly,
}) => {
  const disclaimerText = 'Everyone can see the name of the project.';
  const readOnlyText = 'You need at least projectOwner to edit roles';

  return (
    <RolesModal
      resourceType={ResourceType.PROJECT}
      projectName={projectName}
      disclaimerText={disclaimerText}
      show={show}
      onHide={onHide}
      readOnly={readOnly}
      readOnlyText={readOnlyText}
    />
  );
};

export default ProjectRolesModal;
