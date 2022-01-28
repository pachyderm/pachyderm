import {
  FormModal,
  InfoSVG,
  Link,
  Icon,
  ExternalLinkSVG,
} from '@pachyderm/components';
import React from 'react';

import ScaleQuestion from './components/ScaleQuestion';
import styles from './ExitSurvey.module.css';
import useExitSurvey from './hooks/useExitSurvey';

type ExitSurveyProps = {
  onTutorialClose: () => void;
};

const ExitSurvey: React.FC<ExitSurveyProps> = ({onTutorialClose}) => {
  const {isOpen, onSubmit, formContext, updating, disabled, onClose, error} =
    useExitSurvey(onTutorialClose);

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
          <div className={styles.infoHeader}>How to delete this tutorial</div>
          <div className={styles.infoLine}>
            Delete all repos and pipelines.
            <Link
              className={styles.link}
              small
              externalLink
              to="https://docs.pachyderm.com/latest/reference/pachctl/pachctl_delete_all/"
            >
              Pachctl delete all
              <Icon className={styles.linkSVG} small color="plum">
                <ExternalLinkSVG />
              </Icon>
            </Link>
          </div>
          <div className={styles.infoLine}>
            Select repos and pipelines to delete.
            <Link
              className={styles.link}
              small
              externalLink
              to="https://docs.pachyderm.com/latest/reference/pachctl/pachctl_delete/"
            >
              Pachctl delete all
              <Icon className={styles.linkSVG} small color="plum">
                <ExternalLinkSVG />
              </Icon>
            </Link>
          </div>
        </div>
      </div>
      <div className={styles.card}>
        <div className={styles.headerText}>How was your experience?</div>
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
