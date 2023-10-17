import {SetClusterDefaultsResp} from '@graphqlTypes';
import React, {useState} from 'react';
import {useForm} from 'react-hook-form';

import BrandedDocLink from '@dash-frontend/components/BrandedDocLink';
import {
  FormModal,
  ButtonLink,
  RadioButton,
  CaptionTextSmall,
} from '@pachyderm/components';

import AffectedPipelinesTable from './components/AffectectedPipelinesTable';
import styles from './ConfirmConfigModal.module.css';

type ConfirmConfigModalProps = {
  show: boolean;
  onHide?: () => void;
  onSubmit: ({
    regenerate,
    reprocess,
  }: {
    regenerate: boolean;
    reprocess: boolean;
  }) => void;
  affectedPipelines?: SetClusterDefaultsResp['affectedPipelinesList'];
  loading: boolean;
};

const ConfirmConfigModal: React.FC<ConfirmConfigModalProps> = ({
  show,
  onHide,
  onSubmit,
  affectedPipelines,
  loading,
}) => {
  const [affectedPipelinesOpen, setAffectedPipelinesVisible] = useState(true);
  const formCtx = useForm({
    mode: 'onChange',
    defaultValues: {
      configOption: 'save',
    },
  });

  const {getValues} = formCtx;
  const configOption = getValues('configOption');

  const handleSubmit = () =>
    onSubmit({
      regenerate: ['regenerate', 'reprocess'].includes(configOption),
      reprocess: configOption === 'reprocess',
    });

  return (
    <FormModal
      isOpen={show}
      onHide={onHide}
      confirmText="Save"
      formContext={formCtx}
      headerText="Save Cluster Defaults"
      onSubmit={handleSubmit}
      mode="Long"
      loading={loading}
      footerContent={
        <BrandedDocLink pathWithoutDomain="/set-up/global-config/">
          Info
        </BrandedDocLink>
      }
    >
      <div className={styles.modalContent}>
        <RadioButton id="save" value="save" name="configOption">
          <RadioButton.Label>Save Cluster Defaults</RadioButton.Label>
        </RadioButton>
        <CaptionTextSmall color="black" className={styles.description}>
          Save defaults without regenerating pipeline specs. To use new
          defaults, edit or create a pipeline.
        </CaptionTextSmall>

        <RadioButton id="regenerate" value="regenerate" name="configOption">
          <RadioButton.Label>
            Save Cluster Defaults and Regenerate Pipelines
          </RadioButton.Label>
        </RadioButton>
        <CaptionTextSmall color="black" className={styles.description}>
          Save defaults and regenerate pipeline specs. Only specs changed by the
          new defaults will be regenerated. Previously-processed datums will not
          be reprocessed.
        </CaptionTextSmall>

        <RadioButton id="reprocess" value="reprocess" name="configOption">
          <RadioButton.Label>
            Save Cluster Defaults, Regenerate Pipelines, and Reprocess Datums
          </RadioButton.Label>
        </RadioButton>
        <CaptionTextSmall color="black" className={styles.description}>
          Save defaults, regenerate pipeline specs, and reprocess all datums.
          Only specs changed by the new defaults will be regenerated.
        </CaptionTextSmall>
        {configOption !== 'save' && affectedPipelines && (
          <div className={styles.affectedPipelinesDisclaimer}>
            <span className={styles.disclaimerText}>
              {affectedPipelines.length} pipeline
              {affectedPipelines.length !== 1 ? 's' : ''} will be affected
            </span>
            {affectedPipelines.length > 0 && (
              <ButtonLink
                onClick={() => {
                  setAffectedPipelinesVisible(!affectedPipelinesOpen);
                }}
                type="button"
              >
                {affectedPipelinesOpen ? 'Hide List' : 'Show List'}
              </ButtonLink>
            )}
          </div>
        )}
        {affectedPipelinesOpen &&
          affectedPipelines &&
          affectedPipelines.length > 0 &&
          configOption !== 'save' && (
            <AffectedPipelinesTable affectedPipelines={affectedPipelines} />
          )}
      </div>
    </FormModal>
  );
};

export default ConfirmConfigModal;
