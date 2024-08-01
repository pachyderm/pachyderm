import React from 'react';
import {useForm} from 'react-hook-form';

import {FormModal, Checkbox, CaptionTextSmall} from '@pachyderm/components';

import styles from './ConfirmUpdateModal.module.css';

type ConfirmUpdateModalProps = {
  show: boolean;
  onHide?: () => void;
  onSubmit: ({reprocess}: {reprocess: boolean}) => void;
  loading: boolean;
};

type UpdateFields = {
  reprocess: boolean;
};

const ConfirmUpdateModal: React.FC<ConfirmUpdateModalProps> = ({
  show,
  onHide,
  onSubmit,
  loading,
}) => {
  const formCtx = useForm<UpdateFields>({
    mode: 'onChange',
  });

  const handleSubmit = (values: UpdateFields) => {
    onSubmit({
      reprocess: values.reprocess,
    });
  };

  return (
    <FormModal
      isOpen={show}
      onHide={onHide}
      confirmText="Update Pipeline"
      formContext={formCtx}
      headerText="Update Pipeline"
      onSubmit={handleSubmit}
      loading={loading}
    >
      <div className={styles.modalContent}>
        <div>
          Update this pipeline with new specifications, and run a new subjob to
          process data.
        </div>
        <CaptionTextSmall className={styles.optional}>
          Optional Settings
        </CaptionTextSmall>
        <Checkbox name="reprocess" label="Reprocess Datums" />
        <CaptionTextSmall color="black" className={styles.description}>
          Update this pipeline with new specifications and reprocess all datums.
        </CaptionTextSmall>
      </div>
    </FormModal>
  );
};

export default ConfirmUpdateModal;
