import classnames from 'classnames';
import isEqual from 'lodash/isEqual';
import React, {useEffect, useState} from 'react';
import {useForm} from 'react-hook-form';

import BrandedDocLink from '@dash-frontend/components/BrandedDocLink';
import CodePreview from '@dash-frontend/components/CodePreview';
import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import {
  BasicModal,
  Button,
  ButtonLink,
  ExternalLinkSVG,
  Form,
  Icon,
  Input,
  SpinnerSVG,
  StatusCheckmarkSVG,
  StatusWarningSVG,
  useDebounce,
  useModal,
} from '@pachyderm/components';

import parseTemplateMetadata from '../../lib/parseTemplateMetadata';
import {Template} from '../../templates';

import styles from './UploadTemplate.module.css';

type UploadTemplateProps = {
  selectedTemplate: Template | undefined;
  setSelectedTemplate: React.Dispatch<
    React.SetStateAction<Template | undefined>
  >;
};

type UploadStatus = {
  state: 'success' | 'error';
  message: string;
  fullError?: unknown;
};

const UploadTemplate: React.FC<UploadTemplateProps> = ({
  selectedTemplate,
  setSelectedTemplate,
}) => {
  const {isOpen, openModal, closeModal} = useModal();
  const {enterpriseActive} = useEnterpriseActive();
  const [uploadedTemplate, setUploadedTemplate] = useState<
    Template | undefined
  >(undefined);

  const [validUrl, setValidUrl] = useState<string | undefined>(undefined);
  const [status, setStatus] = useState<UploadStatus | null>();
  const [loading, setLoading] = useState(false);

  const urlCtx = useForm({
    mode: 'onChange',
  });

  const {watch} = urlCtx;
  const url = watch('url');
  const debouncedUrl = useDebounce(url, 200);

  useEffect(() => {
    setStatus(null);
    setUploadedTemplate(undefined);
    setSelectedTemplate(undefined);
    try {
      if (debouncedUrl) {
        // Validate the URL using the constructor.
        new URL(debouncedUrl);
        setValidUrl(debouncedUrl);
      }
    } catch {
      setValidUrl(undefined);
      setStatus({state: 'error', message: 'Invalid url'});
    }
  }, [debouncedUrl, setSelectedTemplate]);

  useEffect(() => {
    const fetchData = async () => {
      let raw;
      if (validUrl) {
        setLoading(true);
        try {
          const response = await fetch(validUrl);
          if (!response.ok) {
            throw new Error();
          }

          raw = await response.text();
        } catch {
          setStatus({
            state: 'error',
            message: 'Error retrieving template from url',
          });
        }
        setLoading(false);
        if (raw) {
          try {
            const metadata = parseTemplateMetadata(raw);
            if (metadata) {
              const template = {
                metadata,
                template: raw,
              };
              setUploadedTemplate(template);
              setSelectedTemplate(template);
              setStatus({
                state: 'success',
                message: 'Template uploaded',
              });
            } else {
              setStatus({
                state: 'error',
                message: 'No template metadata found',
              });
            }
          } catch (e) {
            setStatus({
              state: 'error',
              message: 'Error parsing metadata from template',
              fullError: e,
            });
          }
        }
      }
    };

    fetchData();
  }, [setSelectedTemplate, validUrl]);

  return (
    <>
      {isOpen && (
        <BasicModal
          show={isOpen}
          onHide={closeModal}
          headerContent={'Metadata Parsing Error'}
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
          We have encountered an error parsing the metadata in the uploaded
          template.
          <CodePreview
            source={getErrorMessage(status?.fullError)}
            language="text"
            wrapText
          />
          More information avaliable in our{' '}
          <BrandedDocLink pathWithoutDomain="/build-dags/pipeline-operations/jsonnet-pipeline-specs/#console">
            docs
          </BrandedDocLink>
        </BasicModal>
      )}

      <div
        className={classnames(styles.card, {
          [styles.selected]:
            selectedTemplate && isEqual(uploadedTemplate, selectedTemplate),
          [styles.clickable]: !!uploadedTemplate,
        })}
        onClick={() => {
          if (uploadedTemplate) {
            setSelectedTemplate(uploadedTemplate);
          }
        }}
      >
        <label htmlFor="url">
          <h6>Add template from URL</h6>
        </label>
        <Form formContext={urlCtx}>
          <Input name="url" placeholder="https://" />
        </Form>
        {loading && (
          <Icon small>
            <SpinnerSVG />
          </Icon>
        )}
        {status?.message && (
          <div className={styles.status}>
            {status.state === 'success' ? (
              <Icon small color="green">
                <StatusCheckmarkSVG />
              </Icon>
            ) : (
              <Icon small color="red">
                <StatusWarningSVG />
              </Icon>
            )}
            {status.message}
            {!!status.fullError && (
              <ButtonLink
                small
                onClick={openModal}
                className={styles.fullError}
              >
                See Full Error
              </ButtonLink>
            )}
          </div>
        )}
      </div>
    </>
  );
};

export default UploadTemplate;
