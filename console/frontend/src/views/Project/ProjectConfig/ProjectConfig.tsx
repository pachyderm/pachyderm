import React from 'react';

import CodeEditor from '@dash-frontend/components/CodeEditor';
import CodeEditorInfoBar from '@dash-frontend/components/CodeEditorInfoBar';
import ConfirmConfigModal from '@dash-frontend/components/ConfirmConfigModal';
import View from '@dash-frontend/components/View';
import {
  Group,
  Button,
  ArrowLeftSVG,
  ArrowRightSVG,
  useModal,
  BasicModal,
  SpinnerSVG,
} from '@pachyderm/components';

import useProjectConfig from './hooks/useProjectConfig';
import styles from './ProjectConfig.module.css';

type ProjectConfigProps = {
  triggerNotification?: (
    message: string,
    type: 'success' | 'error',
    duration?: number,
  ) => void;
};

const ProjectConfig: React.FC<ProjectConfigProps> = ({triggerNotification}) => {
  const {
    openModal: openApplyConfigModal,
    closeModal: closeApplyConfigModal,
    isOpen: applyConfigModalOpen,
  } = useModal(false);
  const {
    openModal: openUnsavedChangesModal,
    closeModal: closeUnsavedChangesModal,
    isOpen: unsavedChangesModalOpen,
  } = useModal(false);

  const {
    editorText,
    setEditorText,
    initialEditorDoc,
    error,
    unsavedChanges,
    isValidJSON,
    setUnsavedChanges,
    prettifyJSON,
    getProjectLoading,
    setProjectDefaultsMutation,
    setProjectResponse,
    setProjectLoading,
    goToProject,
    fullError,
    projectId,
  } = useProjectConfig();

  return (
    <View className={styles.view} sidenav={false} canvas={false}>
      <Group justify="stretch" className={styles.titleBar}>
        <h5>Project Defaults</h5>
        <Group spacing={8}>
          <Button
            IconSVG={ArrowLeftSVG}
            buttonType="ghost"
            onClick={() =>
              unsavedChanges ? openUnsavedChangesModal() : goToProject()
            }
          >
            Back
          </Button>
          <Button
            IconSVG={setProjectLoading ? SpinnerSVG : ArrowRightSVG}
            disabled={!isValidJSON || Boolean(error) || setProjectLoading}
            buttonType="primary"
            iconPosition="end"
            onClick={() =>
              setProjectDefaultsMutation(
                {
                  project: {name: projectId},
                  projectDefaultsJson: editorText,
                  regenerate: true,
                  dryRun: true,
                },
                {
                  onSuccess: openApplyConfigModal,
                },
              )
            }
          >
            Continue
          </Button>
        </Group>
      </Group>
      <div className={styles.editor}>
        <CodeEditorInfoBar
          errorMessage={error}
          fullError={fullError}
          fullErrorModalTitle="Project Defaults Error"
          unsavedChanges={unsavedChanges}
          invalidJSON={!isValidJSON}
          handlePrettify={prettifyJSON}
          docsLink="/set-up/global-config/"
        />
        <CodeEditor
          className={styles.codePreview}
          loading={getProjectLoading}
          source={editorText}
          initialDoc={initialEditorDoc}
          onChange={setEditorText}
          schema="projectDefaults"
        />
      </div>

      {applyConfigModalOpen && (
        <ConfirmConfigModal
          level="Project"
          show={applyConfigModalOpen}
          loading={setProjectLoading}
          onHide={closeApplyConfigModal}
          onSubmit={({
            regenerate,
            reprocess,
          }: {
            regenerate: boolean;
            reprocess: boolean;
          }) =>
            setProjectDefaultsMutation(
              {
                project: {
                  name: projectId,
                },
                projectDefaultsJson: editorText,
                regenerate,
                reprocess,
              },
              {
                onSuccess: () => {
                  closeApplyConfigModal();
                  setUnsavedChanges(false);
                  goToProject();
                  triggerNotification &&
                    triggerNotification(
                      'Project defaults saved successfuly',
                      'success',
                    );
                },
              },
            )
          }
          affectedPipelines={setProjectResponse?.affectedPipelines || []}
        />
      )}

      {unsavedChangesModalOpen && (
        <BasicModal
          show={unsavedChangesModalOpen}
          onHide={closeUnsavedChangesModal}
          headerContent="Unsaved Changes"
          actionable
          mode="Small"
          loading={false}
          confirmText="Leave"
          onConfirm={goToProject}
        >
          You have unsaved changes that will be lost. Are you sure you want to
          leave the Project Defaults editor?
        </BasicModal>
      )}
    </View>
  );
};

export default ProjectConfig;
