import {ButtonLink} from '@pachyderm/components';
import React, {useState} from 'react';

import ConnectModal from '@dash-frontend/components/ConnectModal';
import {useWorkspace} from '@dash-frontend/hooks/useWorkspace';

import styles from './EmptyState.module.css';

type EmptyStateProps = {
  title: string;
  message?: string;
  connect?: boolean;
};

const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  message = null,
  connect,
  children = null,
}) => {
  const {workspaceName, pachdAddress, pachVersion} = useWorkspace();
  const [connectModalShow, showConnectModal] = useState(false);

  return (
    <div className={styles.base}>
      <img
        src="/elephant_empty_state.png"
        className={styles.elephantImage}
        alt=""
      />
      <span className={styles.title}>{title}</span>
      <span className={styles.message}>
        {message}
        {children}
      </span>
      {connect && (
        <ButtonLink
          onClick={() => showConnectModal(true)}
          className={styles.message}
        >
          Connect to Pachctl
        </ButtonLink>
      )}
      <ConnectModal
        show={connectModalShow}
        onHide={() => showConnectModal(false)}
        workspaceName={workspaceName || ''}
        pachdAddress={pachdAddress || ''}
        pachVersion={pachVersion || ''}
      />
    </div>
  );
};

export default EmptyState;
