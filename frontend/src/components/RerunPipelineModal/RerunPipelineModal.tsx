import React, {FunctionComponent} from 'react';

import {
  BasicModal,
  CaptionTextSmall,
  Form,
  RadioButton,
} from '@pachyderm/components';
import getServerErrorMessage from 'lib/errorHandling';

import BrandedDocLink from '../BrandedDocLink/BrandedDocLink';

import useRerunPipelineModal from './hooks/useRerunPipelineModal';
import styles from './RerunPipelineModal.module.css';

type RerunPipelineModalProps = {
  show: boolean;
  onHide: () => void;
  pipelineId: string;
};

const RerunPipelineModal: FunctionComponent<RerunPipelineModalProps> = ({
  show,
  onHide,
  pipelineId,
}) => {
  const {
    formCtx,
    onRerunPipeline,
    updating,
    disabled,
    error,
    projectId,
    successMessage,
  } = useRerunPipelineModal(pipelineId, onHide);

  return (
    <BasicModal
      show={show}
      onHide={onHide}
      headerContent={`Rerun Pipeline: ${projectId}/${pipelineId}`}
      actionable
      mode="Default"
      confirmText="Rerun Pipeline"
      onConfirm={onRerunPipeline}
      updating={updating}
      loading={false}
      disabled={disabled}
      successMessage={successMessage}
      errorMessage={getServerErrorMessage(error)}
      footerContent={
        <BrandedDocLink pathWithoutDomain="/run-commands/pachctl_rerun_pipeline/">
          Info
        </BrandedDocLink>
      }
    >
      <div className={styles.base}>
        <Form formContext={formCtx}>
          <RadioButton id="true" name="reprocess" value="true" small>
            <RadioButton.Label>Process All Datums</RadioButton.Label>
          </RadioButton>
          <CaptionTextSmall color="black" className={styles.description}>
            Process all datums including previously successful datums.
          </CaptionTextSmall>
          <RadioButton id="false" name="reprocess" value="false" small>
            <RadioButton.Label>Process only Failed Datums</RadioButton.Label>
          </RadioButton>
          <CaptionTextSmall color="black" className={styles.description}>
            Process only the failed datums from the previous subjob.
          </CaptionTextSmall>
        </Form>
      </div>
    </BasicModal>
  );
};

export default RerunPipelineModal;
