import React from 'react';

import {
  DefaultDropdown,
  ChevronDownSVG,
  OverflowSVG,
  ButtonTypeOptions,
} from '@pachyderm/components';

import DeletePipelineModal from './components/DeletePipelineModal';
import RerunPipelineModal from './components/RerunPipelineModal';
import usePipelineActionsMenu from './hooks/usePipelineActionsMenu';
import styles from './PipelineActionsMenu.module.css';

type PipelineActionsModalsProps = {
  pipelineId: string;
  closeRerunPipelineModal: () => void;
  rerunPipelineModalIsOpen: boolean;
  deletePipelineModalOpen: boolean;
  setDeleteModalOpen: React.Dispatch<React.SetStateAction<boolean>>;
};

export const PipelineActionsModals: React.FC<PipelineActionsModalsProps> = ({
  pipelineId,
  closeRerunPipelineModal,
  rerunPipelineModalIsOpen,
  deletePipelineModalOpen,
  setDeleteModalOpen,
}) => {
  return (
    <>
      {rerunPipelineModalIsOpen && (
        <RerunPipelineModal
          pipelineId={pipelineId}
          show={rerunPipelineModalIsOpen}
          onHide={closeRerunPipelineModal}
        />
      )}
      {deletePipelineModalOpen && (
        <DeletePipelineModal
          setModalOpen={setDeleteModalOpen}
          pipelineId={pipelineId}
        />
      )}
    </>
  );
};

type PipelineActionsMenuProps = {
  pipelineId: string;
};

const PipelineActionsMenu: React.FC<PipelineActionsMenuProps> = ({
  pipelineId,
}) => {
  const {
    closeRerunPipelineModal,
    rerunPipelineModalIsOpen,
    deletePipelineModalOpen,
    setDeleteModalOpen,
    onDropdownMenuSelect,
    menuItems,
    onMenuOpen,
  } = usePipelineActionsMenu(pipelineId);

  const dropdownClasses = {
    items: menuItems,
    onSelect: onDropdownMenuSelect,
    menuOpts: {pin: 'right' as 'right' | undefined},
    openOnClick: onMenuOpen,
  };

  const buttonOpts = {
    hideChevron: true,
    buttonType: 'secondary' as ButtonTypeOptions,
    color: 'purple',
  };

  return (
    <div className={styles.dropdownWrapper}>
      <DefaultDropdown
        className={styles.fullContent}
        buttonOpts={{
          IconSVG: ChevronDownSVG,
          ...buttonOpts,
        }}
        {...dropdownClasses}
      >
        Pipeline Actions
      </DefaultDropdown>
      <DefaultDropdown
        className={styles.responsiveContent}
        buttonOpts={{
          IconSVG: OverflowSVG,
          ...buttonOpts,
        }}
        aria-label="pipeline-actions-menu"
        {...dropdownClasses}
      />
      <PipelineActionsModals
        pipelineId={pipelineId}
        closeRerunPipelineModal={closeRerunPipelineModal}
        rerunPipelineModalIsOpen={rerunPipelineModalIsOpen}
        deletePipelineModalOpen={deletePipelineModalOpen}
        setDeleteModalOpen={setDeleteModalOpen}
      />
    </div>
  );
};

export default PipelineActionsMenu;
