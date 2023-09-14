import {ApolloError} from '@apollo/client';
import React, {useEffect} from 'react';

import {
  ErrorText,
  StatusWarningSVG,
  Icon,
  BasicModal,
  useModal,
} from '@pachyderm/components';

import styles from './DAGError.module.css';

const LINEAGE_ERROR = 'Unable to construct DAG';
const CONSTRUCTION_ERROR_DETAILS =
  'Unable to construct DAG from repos and pipelines. You may want to look into your pipeline inputs.';
const NETWORK_ERROR_DETAILS =
  'Repo and Pipeline data may not be up to date. You may want to refresh the page.';

type DAGErrorProps = {
  error: ApolloError | string | undefined;
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

  const dagBuildError =
    typeof error === 'string' && error.includes(LINEAGE_ERROR);

  return (
    <>
      {error && (
        <div className={styles.errorBox}>
          <Icon small color="red">
            <StatusWarningSVG />
          </Icon>
          <ErrorText className={styles.dagError}>
            {dagBuildError
              ? error
              : 'Connection error: data may not be up to date.'}
          </ErrorText>
        </div>
      )}
      <BasicModal
        show={errorModalOpen}
        headerContent={
          <span className={styles.titleError}>
            <Icon small color="red">
              <StatusWarningSVG />
            </Icon>
            <h4>{dagBuildError ? ' Lineage Error' : ' Connection Error'}</h4>
          </span>
        }
        mode="Small"
        loading={false}
        onHide={closeErrorModal}
      >
        {dagBuildError ? CONSTRUCTION_ERROR_DETAILS : NETWORK_ERROR_DETAILS}
      </BasicModal>
    </>
  );
};

export default DAGError;
