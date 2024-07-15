import React, {FunctionComponent} from 'react';

import {
  BasicModal,
  CaptionTextSmall,
  Form,
  RadioButton,
} from '@pachyderm/components';

import BrandedDocLink from '../../../BrandedDocLink/BrandedDocLink';

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
  const {formCtx, onRerunPipeline, updating, disabled, error, projectId} =
    useRerunPipelineModal(pipelineId, onHide);

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
      errorMessage={error}
      footerContent={
        <BrandedDocLink pathWithoutDomain="/run-commands/pachctl_rerun_pipeline/">
          Info
        </BrandedDocLink>
      }
    >
      <div className={styles.base}>
        <div className={styles.explanation}>
          Rerunning a pipeline will increment the version number of the
          specification.
        </div>
        <Form formContext={formCtx}>
          <RadioButton id="true" name="reprocess" value="true" small>
            <RadioButton.Label>Reprocess all datums</RadioButton.Label>
          </RadioButton>
          <CaptionTextSmall color="black" className={styles.description}>
            All datums from the previous job will be reprocessed. This could
            result in more processing than just reprocessing failed datums.
          </CaptionTextSmall>
          <RadioButton id="false" name="reprocess" value="false" small>
            <RadioButton.Label>Reprocess only failed datums</RadioButton.Label>
          </RadioButton>
          <CaptionTextSmall color="black" className={styles.description}>
            Only process failed datums from the previous job, all others will
            appear skipped.
          </CaptionTextSmall>
        </Form>
      </div>
    </BasicModal>
  );
};

export default RerunPipelineModal;
