import React, {FunctionComponent} from 'react';

import {ResourceType} from '@dash-frontend/api/auth';

import {RolesModal} from '../RolesModal';

type ClusterRolesModalProps = {
  show: boolean;
  onHide?: () => void;
  readOnly?: boolean;
};

const ClusterRolesModal: FunctionComponent<ClusterRolesModalProps> = ({
  show,
  onHide,
  readOnly,
}) => {
  const readOnlyText = 'You need to be a cluster admin to edit roles';

  return (
    <RolesModal
      resourceType={ResourceType.CLUSTER}
      show={show}
      onHide={onHide}
      readOnly={readOnly}
      readOnlyText={readOnlyText}
    />
  );
};

export default ClusterRolesModal;
