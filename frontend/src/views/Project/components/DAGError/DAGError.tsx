import {ApolloError} from '@apollo/client';
import {
  ErrorText,
  StatusWarningSVG,
  Icon,
  BasicModal,
  useModal,
} from '@pachyderm/components';
import React, {useEffect} from 'react';

import styles from './DAGError.module.css';

type DAGErrorProps = {
  error: ApolloError | undefined;
};

const DAGError: React.FC<DAGErrorProps> = ({error}) => {
  const {
    closeModal: closeErrorModal,
    openModal: openErrorModal,
    isOpen: errorModalOpen,
  } = useModal(false);

  useEffect(() => {
    if (error) {
      openErrorModal();
    }
  }, [error, openErrorModal]);

  return (
    <>
      {error && (
        <div className={styles.errorBox}>
          <Icon small>
            <StatusWarningSVG />
          </Icon>
          <ErrorText className={styles.dagError}>
            Connection error: data may not be up to date.
          </ErrorText>
        </div>
      )}
      <BasicModal
        show={errorModalOpen}
        headerContent={
          <>
            <StatusWarningSVG />
            {` Connection Error`}
          </>
        }
        small
        loading={false}
        onHide={closeErrorModal}
      >
        Repo and Pipeline data may not be up to date. You may want to refresh
        the page.
      </BasicModal>
    </>
  );
};

export default DAGError;
