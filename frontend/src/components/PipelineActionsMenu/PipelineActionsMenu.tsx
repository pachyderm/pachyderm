import React from 'react';

import {DefaultDropdown, ChevronDownSVG} from '@pachyderm/components';

import DeletePipelineModal from './components/DeletePipelineModal';
import RerunPipelineModal from './components/RerunPipelineModal';
import usePipelineActionsMenu from './hooks/usePipelineActionsMenu';

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

  return (
    <>
      <DefaultDropdown
        items={menuItems}
        onSelect={onDropdownMenuSelect}
        buttonOpts={{
          hideChevron: true,
          buttonType: 'secondary',
          IconSVG: ChevronDownSVG,
          color: 'purple',
        }}
        menuOpts={{pin: 'right'}}
        openOnClick={onMenuOpen}
      >
        Pipeline Actions
      </DefaultDropdown>
      <PipelineActionsModals
        pipelineId={pipelineId}
        closeRerunPipelineModal={closeRerunPipelineModal}
        rerunPipelineModalIsOpen={rerunPipelineModalIsOpen}
        deletePipelineModalOpen={deletePipelineModalOpen}
        setDeleteModalOpen={setDeleteModalOpen}
      />
    </>
  );
};

export default PipelineActionsMenu;
