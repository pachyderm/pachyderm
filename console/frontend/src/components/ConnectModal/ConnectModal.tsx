import React, {memo} from 'react';

import {BasicModal, Group, Login, Connect, Verify} from '@pachyderm/components';

import styles from './ConnectModal.module.css';

export type ConnectModalProps = {
  show: boolean;
  onHide: () => void;
  workspaceName: string;
  pachdAddress: string;
  pachVersion: string;
};

const ConnectModal: React.FC<ConnectModalProps> = ({
  onHide,
  show,
  workspaceName,
  pachdAddress,
  pachVersion,
}) => {
  return (
    <BasicModal
      data-testid="ConnectModal__basicModal"
      show={show}
      onHide={onHide}
      className={styles.modal}
      loading={false}
      headerContent={<span>Connect to Workspace</span>}
      hideActions
    >
      <ol className={styles.descriptionList}>
        <Group vertical spacing={32}>
          <Connect name={workspaceName} pachdAddress={pachdAddress} />
          <Login hasOTP={false} />
          <Verify>
            <tr>
              <td>pachctl</td>
              <td data-testid="ConnectModal__pachctlVersion">{pachVersion}</td>
            </tr>
            <tr>
              <td>pachd</td>
              <td data-testid="ConnectModal__pachdVersion">{pachVersion}</td>
            </tr>
          </Verify>
        </Group>
      </ol>
    </BasicModal>
  );
};

export default memo(ConnectModal);
