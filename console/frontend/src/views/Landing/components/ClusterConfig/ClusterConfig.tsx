import React from 'react';

import CodeEditor from '@dash-frontend/components/CodeEditor';
import CodeEditorInfoBar from '@dash-frontend/components/CodeEditorInfoBar';
import ConfirmConfigModal from '@dash-frontend/components/ConfirmConfigModal';
import View from '@dash-frontend/components/View';
import EditMetadataModal from '@dash-frontend/views/Project/components/ProjectSidebar/components/RepoDetails/components/UserMetadata/components/EditMetadataModal';
import {
  Group,
  Button,
  ArrowLeftSVG,
  ArrowRightSVG,
  useModal,
  BasicModal,
  SpinnerSVG,
} from '@pachyderm/components';

import styles from './ClusterConfig.module.css';
import useClusterConfig from './hooks/useClusterConfig';

type ClusterConfigProps = {
  triggerNotification?: (
    message: string,
    type: 'success' | 'error',
    duration?: number,
  ) => void;
};

const ClusterConfig: React.FC<ClusterConfigProps> = ({triggerNotification}) => {
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
    getClusterLoading,
    setClusterDefaultsMutation,
    setClusterResponse,
    setClusterLoading,
    goToLanding,
    fullError,
    isMetadataOpen,
    openMetadataModal,
    closeMetadataModal,
    clusterMetadataArray,
    hasClusterEditMetadata,
  } = useClusterConfig();

  return (
    <View className={styles.view} sidenav={false} canvas={false}>
      <Group justify="stretch" className={styles.titleBar}>
        <h5>Cluster Defaults</h5>
        <Group spacing={8}>
          <Button
            IconSVG={ArrowLeftSVG}
            buttonType="ghost"
            onClick={() =>
              unsavedChanges ? openUnsavedChangesModal() : goToLanding()
            }
          >
            Back
          </Button>
          {hasClusterEditMetadata && (
            <Button
              buttonType="secondary"
              onClick={(e) => {
                e.stopPropagation();
                openMetadataModal();
              }}
            >
              Cluster Metadata
            </Button>
          )}
          {hasClusterEditMetadata && isMetadataOpen && (
            <EditMetadataModal
              metadata={clusterMetadataArray}
              metadataType="cluster"
              closeModal={closeMetadataModal}
            />
          )}
          <Button
            IconSVG={setClusterLoading ? SpinnerSVG : ArrowRightSVG}
            disabled={!isValidJSON || Boolean(error) || setClusterLoading}
            buttonType="primary"
            iconPosition="end"
            onClick={() =>
              setClusterDefaultsMutation(
                {
                  clusterDefaultsJson: editorText,
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
          fullErrorModalTitle="Cluster Defaults Error"
          unsavedChanges={unsavedChanges}
          invalidJSON={!isValidJSON}
          handlePrettify={prettifyJSON}
          docsLink="/set-up/global-config/"
        />
        <CodeEditor
          className={styles.codePreview}
          loading={getClusterLoading}
          source={editorText}
          initialDoc={initialEditorDoc}
          onChange={setEditorText}
          schema="clusterDefaults"
        />
      </div>

      {applyConfigModalOpen && (
        <ConfirmConfigModal
          level="Cluster"
          show={applyConfigModalOpen}
          loading={setClusterLoading}
          onHide={closeApplyConfigModal}
          onSubmit={({
            regenerate,
            reprocess,
          }: {
            regenerate: boolean;
            reprocess: boolean;
          }) =>
            setClusterDefaultsMutation(
              {
                clusterDefaultsJson: editorText,
                regenerate,
                reprocess,
              },
              {
                onSuccess: () => {
                  closeApplyConfigModal();
                  setUnsavedChanges(false);
                  goToLanding();
                  triggerNotification &&
                    triggerNotification(
                      'Cluster defaults saved successfuly',
                      'success',
                    );
                },
              },
            )
          }
          affectedPipelines={setClusterResponse?.affectedPipelines || []}
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
          onConfirm={goToLanding}
        >
          You have unsaved changes that will be lost. Are you sure you want to
          leave the Cluster Defaults editor?
        </BasicModal>
      )}
    </View>
  );
};

export default ClusterConfig;
