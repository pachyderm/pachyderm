import capitalize from 'lodash/capitalize';
import React from 'react';

import BrandedDocLink from '@dash-frontend/components/BrandedDocLink';
import CodePreview from '@dash-frontend/components/CodePreview';
import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import getErrorMessage, {getGRPCCode} from '@dash-frontend/lib/getErrorMessage';
import {
  Group,
  Icon,
  StatusWarningSVG,
  ButtonLink,
  useModal,
  BasicModal,
  ExternalLinkSVG,
  Button,
  ErrorText,
} from '@pachyderm/components';

import styles from './CodeEditorInfoBar.module.css';

type CodeEditorInfoBarProps = {
  errorMessage?: string;
  fullError?: Error | null;
  fullErrorModalTitle?: string;
  unsavedChanges?: boolean;
  invalidJSON?: boolean;
  handlePrettify?: () => void;
  docsLink?: string;
};

const CodeEditorInfoBar: React.FC<CodeEditorInfoBarProps> = ({
  errorMessage,
  fullError,
  fullErrorModalTitle,
  unsavedChanges,
  invalidJSON,
  handlePrettify,
  docsLink,
}) => {
  const {isOpen, openModal, closeModal} = useModal();
  const {enterpriseActive} = useEnterpriseActive();

  const grpcCode = getGRPCCode(fullError);

  return (
    <Group spacing={16} className={styles.infoBar}>
      <Group spacing={8} className={styles.error}>
        {invalidJSON && (
          <>
            <Icon small color="red">
              <StatusWarningSVG />
            </Icon>
            <span role="alert">Invalid JSON</span>
          </>
        )}
        {!invalidJSON && errorMessage && (
          <>
            <Icon small color="red">
              <StatusWarningSVG />
            </Icon>
            <span role="alert">{capitalize(errorMessage)}</span>
            {fullError && (
              <ButtonLink
                small
                onClick={openModal}
                className={styles.fullError}
              >
                See Full Error{' '}
              </ButtonLink>
            )}
          </>
        )}
        {!errorMessage && !invalidJSON && unsavedChanges && (
          <>
            <Icon small color="yellow">
              <StatusWarningSVG />
            </Icon>
            <span role="alert">You have unsaved changes</span>
          </>
        )}
      </Group>
      {handlePrettify && (
        <ButtonLink onClick={handlePrettify}>Prettify</ButtonLink>
      )}
      {docsLink && (
        <BrandedDocLink pathWithoutDomain={docsLink}>Learn More</BrandedDocLink>
      )}

      {isOpen && fullError && (
        <BasicModal
          show={isOpen}
          onHide={closeModal}
          headerContent={fullErrorModalTitle || 'Full Error'}
          mode="Long"
          loading={false}
          actionable
          hideConfirm
          cancelText="Back"
          footerContent={
            <Button
              onClick={() =>
                window.open(enterpriseActive ? EMAIL_SUPPORT : SLACK_SUPPORT)
              }
              IconSVG={ExternalLinkSVG}
              iconPosition="end"
              buttonType="ghost"
            >
              Contact Support
            </Button>
          }
        >
          {grpcCode && (
            <div>
              <ErrorText>{grpcCode}</ErrorText>
            </div>
          )}
          {capitalize(errorMessage)}
          <CodePreview
            className={styles.fullError}
            source={getErrorMessage(fullError)}
            language="json"
            wrapText
          />
        </BasicModal>
      )}
    </Group>
  );
};

export default CodeEditorInfoBar;
