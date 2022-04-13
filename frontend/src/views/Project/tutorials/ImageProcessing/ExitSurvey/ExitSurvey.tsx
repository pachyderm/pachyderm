import {FormModal, InfoSVG, Description} from '@pachyderm/components';
import React from 'react';

import useAccount from '@dash-frontend/hooks/useAccount';

import ScaleQuestion from './components/ScaleQuestion';
import styles from './ExitSurvey.module.css';
import useExitSurvey from './hooks/useExitSurvey';

type ExitSurveyProps = {
  onTutorialClose: () => void;
};

const ExitSurvey: React.FC<ExitSurveyProps> = ({onTutorialClose}) => {
  const {isOpen, onSubmit, formContext, updating, disabled, onClose, error} =
    useExitSurvey(onTutorialClose);

  const {tutorialId} = useAccount();

  const deleteCommand = `pachctl delete pipeline montage_${tutorialId} && pachctl delete pipeline edges_${tutorialId} && pachctl delete repo images_${tutorialId}`;

  return (
    <FormModal
      isOpen={isOpen}
      loading={false}
      updating={updating}
      formContext={formContext}
      onSubmit={onSubmit}
      confirmText="Submit Survey"
      headerText="Tutorial Complete!"
      disabled={disabled}
      onHide={onClose}
      error={error ? "We weren't able to submit your survey" : undefined}
    >
      <div className={styles.info}>
        <div className={styles.infoIcon}>
          <InfoSVG />
        </div>
        <div className={styles.infoContent}>
          <div className={styles.infoHeader}>
            <h6>How to delete this tutorial</h6>
          </div>
          <div className={styles.infoLine}>
            <Description
              id="delete-these"
              title="Delete repos and pipelines created by this tutorial"
              asListItem={false}
              copyText={deleteCommand}
            >
              {deleteCommand}
            </Description>
          </div>
        </div>
      </div>
      <div className={styles.card}>
        <div className={styles.headerText}>
          <h6>How was your experience?</h6>
        </div>
        <div className={styles.subheaderText}>
          Help us improve by completing this two question survey.
        </div>
        <div className={styles.survey}>
          <ScaleQuestion
            name="exit_survey___q1"
            question="1. I believe that the format of this tutorial is engaging and easy to follow."
          />
          <ScaleQuestion
            name="exit_survey___q2"
            question="2. I would like more tutorials covering different topics and use cases."
          />
        </div>
      </div>
    </FormModal>
  );
};

export default ExitSurvey;
