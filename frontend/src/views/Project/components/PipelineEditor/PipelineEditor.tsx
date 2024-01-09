import React from 'react';

import CodeEditor from '@dash-frontend/components/CodeEditor';
import CodeEditorInfoBar from '@dash-frontend/components/CodeEditorInfoBar';
import View from '@dash-frontend/components/View';
import {
  Group,
  Button,
  ArrowLeftSVG,
  ArrowRightSVG,
  SpinnerSVG,
  EditSVG,
  Tabs,
  Icon,
  NavigationSidebarSVG,
  Tooltip,
  useModal,
  BasicModal,
  PanelLeftSVG,
} from '@pachyderm/components';

import ProjectHeader from '../ProjectHeader';

import CodePreviewTabs from './components/CodePreviewTabs';
import ConfirmUpdateModal from './components/ConfirmUpdateModal';
import usePipelineEditor from './hooks/usePipelineEditor';
import styles from './PipelineEditor.module.css';

const PipelineEditor = () => {
  const {
    openModal: openUpdateModal,
    closeModal: closeUpdateModal,
    isOpen: updateModalOpen,
  } = useModal(false);
  const {
    openModal: openWarningModal,
    closeModal: closeWarningModal,
    isOpen: warningModalOpen,
  } = useModal(false);

  const {
    editorText,
    setEditorText,
    initialDoc,
    error,
    createPipeline,
    createLoading,
    goToLineage,
    prettifyJSON,
    isValidJSON,
    effectiveSpec,
    effectiveSpecLoading,
    clusterDefaults,
    clusterDefaultsJSON,
    clusterDefaultsLoading,
    projectDefaults,
    projectDefaultsJSON,
    projectDefaultsLoading,
    pipelineLoading,
    isCreatingOrDuplicating,
    splitView,
    setSplitView,
    editorTextJSON,
    fullError,
    isChangingProject,
    isChangingName,
  } = usePipelineEditor(closeUpdateModal);

  const continueWithCreateOrUpdate = () => {
    if (isCreatingOrDuplicating) {
      createPipeline({
        createPipelineRequestJson: editorText,
        update: !isCreatingOrDuplicating,
      });
    } else {
      openUpdateModal();
    }
    closeWarningModal();
  };

  const createButtonClick = () => {
    if (isChangingProject || isChangingName) {
      openWarningModal();
    } else {
      continueWithCreateOrUpdate();
    }
  };

  const getCreateButtonSVG = () => {
    if (isCreatingOrDuplicating) {
      return createLoading ? SpinnerSVG : undefined;
    } else {
      return ArrowRightSVG;
    }
  };

  const getWarningModalTitle = () => {
    let title = 'Unexpected ';

    if (isChangingProject) title += 'Project';
    if (isChangingProject && isChangingName) title += ' & ';
    if (isChangingName) title += 'Pipeline';
    title += ' Name';
    if (isChangingProject && isChangingName) title += 's';

    return title;
  };

  const CodePreviewHeaders = () => (
    <>
      <Tabs.Tab id="effective-spec">Effective Spec</Tabs.Tab>
      <Tabs.Tab id="cluster-defaults">Cluster Defaults</Tabs.Tab>
      <Tabs.Tab id="project-defaults">Project Defaults</Tabs.Tab>
    </>
  );

  return (
    <View className={styles.view} sidenav={false} canvas={false}>
      <ProjectHeader />
      <Group justify="stretch" className={styles.titleBar}>
        <h5>
          {isCreatingOrDuplicating ? 'Create Pipeline' : 'Update Pipeline'}
        </h5>
        <Group spacing={8}>
          <Button
            IconSVG={ArrowLeftSVG}
            buttonType="ghost"
            onClick={goToLineage}
          >
            Cancel
          </Button>
          <Button
            IconSVG={getCreateButtonSVG()}
            disabled={!isValidJSON || Boolean(error) || createLoading}
            buttonType="primary"
            iconPosition="end"
            onClick={createButtonClick}
          >
            {isCreatingOrDuplicating ? 'Create Pipeline' : 'Update Pipeline'}
          </Button>
        </Group>
      </Group>
      <div className={styles.editor}>
        <div className={splitView ? styles.editorTabsSplit : styles.editorTabs}>
          <Tabs
            initialActiveTabId="pipeline-spec"
            resetToInitialTabId={splitView}
          >
            <div className={styles.header}>
              <Tabs.TabsHeader className={styles.editorTabHeader}>
                <Tabs.Tab id="pipeline-spec">
                  <Icon small color="plum">
                    <EditSVG />
                  </Icon>
                  Pipeline Spec
                </Tabs.Tab>
                {!splitView && <CodePreviewHeaders />}
              </Tabs.TabsHeader>
              {!splitView && (
                <Tooltip tooltipText="View windows side by side">
                  <Button
                    buttonType="secondary"
                    IconSVG={NavigationSidebarSVG}
                    onClick={() => setSplitView(true)}
                    aria-label="view windows side by side"
                  />
                </Tooltip>
              )}
            </div>

            <Tabs.TabPanel id="pipeline-spec" className={styles.editorTabPanel}>
              <CodeEditorInfoBar
                errorMessage={error}
                handlePrettify={prettifyJSON}
                docsLink="/build-dags/pipeline-spec/"
                invalidJSON={!isValidJSON}
                fullError={fullError}
                fullErrorModalTitle={
                  isCreatingOrDuplicating
                    ? 'Create Pipeline Error'
                    : 'Update Pipeline Error'
                }
              />
              <CodeEditor
                className={styles.codeEditor}
                loading={pipelineLoading}
                source={editorText}
                initialDoc={initialDoc}
                onChange={setEditorText}
                schema="createPipelineRequest"
                dataTestid="CodeEditor__createPipeline"
              />
            </Tabs.TabPanel>
            {!splitView && (
              <CodePreviewTabs
                editorTextJSON={editorTextJSON}
                clusterDefaultsJSON={clusterDefaultsJSON}
                projectDefaultsJSON={projectDefaultsJSON}
                effectiveSpec={effectiveSpec}
                effectiveSpecLoading={effectiveSpecLoading}
                clusterDefaults={clusterDefaults}
                clusterDefaultsLoading={clusterDefaultsLoading}
                projectDefaults={projectDefaults}
                projectDefaultsLoading={projectDefaultsLoading}
              />
            )}
          </Tabs>
        </div>
        {splitView && (
          <div className={styles.editorTabsSplit}>
            <Tabs initialActiveTabId="effective-spec">
              <div className={styles.header}>
                <Tabs.TabsHeader className={styles.editorTabHeader}>
                  <CodePreviewHeaders />
                </Tabs.TabsHeader>
                <Tooltip tooltipText="View a single window">
                  <Button
                    buttonType="secondary"
                    IconSVG={PanelLeftSVG}
                    onClick={() => setSplitView(false)}
                    aria-label="view single window"
                  />
                </Tooltip>
              </div>
              <CodePreviewTabs
                editorTextJSON={editorTextJSON}
                clusterDefaultsJSON={clusterDefaultsJSON}
                projectDefaultsJSON={projectDefaultsJSON}
                effectiveSpec={effectiveSpec}
                effectiveSpecLoading={effectiveSpecLoading}
                clusterDefaults={clusterDefaults}
                clusterDefaultsLoading={clusterDefaultsLoading}
                projectDefaults={projectDefaults}
                projectDefaultsLoading={projectDefaultsLoading}
              />
            </Tabs>
          </div>
        )}
      </div>

      {updateModalOpen && (
        <ConfirmUpdateModal
          show={updateModalOpen}
          loading={createLoading}
          onHide={closeUpdateModal}
          onSubmit={({reprocess}: {reprocess: boolean}) =>
            createPipeline({
              createPipelineRequestJson: editorText,
              update: true,
              reprocess,
            })
          }
        />
      )}

      {warningModalOpen && (
        <BasicModal
          show={warningModalOpen}
          onHide={closeWarningModal}
          headerContent={getWarningModalTitle()}
          actionable
          loading={false}
          mode="Small"
          cancelText="Go Back"
          confirmText="Continue Anyway"
          onConfirm={continueWithCreateOrUpdate}
        >
          <Group vertical spacing={16}>
            {isChangingProject && (
              <div>
                The project <b>{isChangingProject}</b> does not match the
                current project. This will create the pipeline in{' '}
                <b>{isChangingProject}</b>.
              </div>
            )}
            {isChangingName && (
              <div>
                Pipeline names are immutable. Submitting this spec will create a
                new pipeline named <b>{isChangingName}</b>.
              </div>
            )}
          </Group>
        </BasicModal>
      )}
    </View>
  );
};

export default PipelineEditor;
