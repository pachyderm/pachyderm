import React from 'react';

import {BasicModal, Description} from '@pachyderm/components';

type ModalProps = {
  show: boolean;
  onHide?: () => void;
  projectName?: string;
};

const ActiveProjectModal: React.FC<ModalProps> = ({
  show,
  onHide,
  projectName,
}) => {
  const setActiveCommand = `pachctl config update context --project ${projectName}`;

  return (
    <BasicModal
      show={show}
      onHide={onHide}
      loading={false}
      headerContent={`Set Active Project: "${projectName}"`}
      hideActions
    >
      <p>
        In order to begin working on a project other than the default project,
        you must assign it to a pachCTL context. This enables you to safely add
        or update resources such as pipelines and repos without affecting other
        projects.
      </p>
      <Description
        id="setActive"
        copyText={setActiveCommand}
        title="Run the following in your terminal:"
      >
        {setActiveCommand}
      </Description>
    </BasicModal>
  );
};

export default ActiveProjectModal;
