import classnames from 'classnames';
import isEqual from 'lodash/isEqual';
import React from 'react';

import CodeEditorInfoBar from '@dash-frontend/components/CodeEditorInfoBar';
import View from '@dash-frontend/components/View';
import {
  ArrowLeftSVG,
  ArrowRightSVG,
  Button,
  FieldText,
  Form,
  Group,
  Input,
  SpinnerSVG,
} from '@pachyderm/components';

import ProjectHeader from '../ProjectHeader';

import UploadTemplate from './components/UploadTemplate';
import usePipelineTemplate, {sortArguments} from './hooks/usePipelineTemplate';
import styles from './PipelineTemplate.module.css';

const PipelineTemplate = () => {
  const {
    goToDAG,
    stepTwo,
    setStepTwo,
    selectedTemplate,
    setSelectedTemplate,
    formCtx,
    isValid,
    handleSubmit,
    error,
    fullError,
    templates,
    loading,
  } = usePipelineTemplate();

  return (
    <View className={styles.view} sidenav={false} canvas={false}>
      <ProjectHeader />
      <Group justify="stretch" className={styles.titleBar}>
        <h5>Pipeline from Template</h5>
        <Group spacing={8}>
          <Button
            IconSVG={ArrowLeftSVG}
            buttonType="ghost"
            onClick={() => {
              stepTwo ? setStepTwo(false) : goToDAG();
            }}
          >
            {stepTwo ? 'Back' : 'Cancel'}
          </Button>
          <Button
            IconSVG={loading ? SpinnerSVG : ArrowRightSVG}
            iconPosition="end"
            disabled={stepTwo ? !isValid || loading : !selectedTemplate}
            onClick={() => {
              stepTwo ? handleSubmit() : setStepTwo(true);
            }}
          >
            {stepTwo ? 'Create Pipeline' : 'Continue'}
          </Button>
        </Group>
      </Group>
      {stepTwo && selectedTemplate ? (
        <Form formContext={formCtx}>
          <div className={styles.formBase}>
            {error && (
              <CodeEditorInfoBar
                errorMessage={error}
                docsLink={
                  error.indexOf('template') !== -1
                    ? '/build-dags/pipeline-operations/jsonnet-pipeline-specs/'
                    : '/build-dags/pipeline-spec/'
                }
                fullError={fullError}
                fullErrorModalTitle={
                  error.indexOf('template') !== -1
                    ? 'Template Error'
                    : 'Create Pipeline Error'
                }
              />
            )}
            <div className={styles.formHeader}>
              <h6>{selectedTemplate.metadata.title}</h6>
              <div>{selectedTemplate.metadata.description}</div>
            </div>
            <h5>Arguments</h5>
            <div className={styles.formFields}>
              {sortArguments(selectedTemplate.metadata.args).map(
                (argument, index) => (
                  <div
                    key={index}
                    data-testid={`PipelineTemplateField__${argument.name}`}
                  >
                    <label htmlFor={argument.name} className={styles.formLabel}>
                      <FieldText>{argument.name}</FieldText>
                    </label>
                    <p>{argument.description}</p>
                    <Input
                      id={argument.name}
                      key={argument.name}
                      required
                      {...formCtx.register(argument.name, {
                        required: true,
                      })}
                      type={argument.type}
                      placeholder={`Required ${argument.type}`}
                      defaultValue={argument.default}
                    />
                  </div>
                ),
              )}
            </div>
          </div>
        </Form>
      ) : (
        <div className={styles.grid}>
          <div className={styles.urlRow}>
            <UploadTemplate
              selectedTemplate={selectedTemplate}
              setSelectedTemplate={setSelectedTemplate}
            />
          </div>

          {templates.map((template, index) => (
            <div
              className={classnames(styles.card, {
                [styles.selected]: isEqual(template, selectedTemplate),
              })}
              key={index}
              onClick={() => setSelectedTemplate(templates[index])}
            >
              <h6 className={styles.cardTitle}>{template.metadata.title}</h6>
              <div className={styles.description}>
                {template.metadata.description}
              </div>
            </div>
          ))}
        </div>
      )}
    </View>
  );
};

export default PipelineTemplate;
