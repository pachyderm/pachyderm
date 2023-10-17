import {ApolloError} from '@apollo/client';
import capitalize from 'lodash/capitalize';
import React from 'react';

import BrandedDocLink from '@dash-frontend/components/BrandedDocLink';
import CodePreview from '@dash-frontend/components/CodePreview';
import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import getServerErrorMessage, {
  getGRPCCode,
} from '@dash-frontend/lib/errorHandling';
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
  error?: ApolloError | string;
  fullError?: ApolloError;
  fullErrorModalTitle?: string;
  unsavedChanges?: boolean;
  invalidJSON?: boolean;
  handlePrettify?: () => void;
  docsLink?: string;
};

const CodeEditorInfoBar: React.FC<CodeEditorInfoBarProps> = ({
  error,
  fullError,
  fullErrorModalTitle,
  unsavedChanges,
  invalidJSON,
  handlePrettify,
  docsLink,
}) => {
  const {isOpen, openModal, closeModal} = useModal();
  const {enterpriseActive} = useEnterpriseActive();

  const readableError = capitalize(
    String(typeof error === 'string' ? error : getServerErrorMessage(error)),
  );
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
        {!invalidJSON && error && (
          <>
            <Icon small color="red">
              <StatusWarningSVG />
            </Icon>
            <span role="alert">{readableError}</span>
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
        {!error && !invalidJSON && unsavedChanges && (
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
          {readableError}
          <CodePreview
            className={styles.fullError}
            source={getServerErrorMessage(fullError)}
            language="json"
            wrapText
          />
        </BasicModal>
      )}
    </Group>
  );
};

export default CodeEditorInfoBar;
