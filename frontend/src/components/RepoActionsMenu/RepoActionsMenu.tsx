import React from 'react';

import {DefaultDropdown, ChevronDownSVG} from '@pachyderm/components';

import DeleteRepoModal from './components/DeleteRepoModal';
import useRepoActionsMenu from './hooks/useRepoActionsMenu';

type RepoActionsModalProps = {
  repoId: string;
  deleteRepoModalOpen: boolean;
  setDeleteModalOpen: React.Dispatch<React.SetStateAction<boolean>>;
};

export const RepoActionsModals: React.FC<RepoActionsModalProps> = ({
  repoId,
  deleteRepoModalOpen,
  setDeleteModalOpen,
}) => {
  return (
    <>
      {deleteRepoModalOpen && (
        <DeleteRepoModal repoId={repoId} setModalOpen={setDeleteModalOpen} />
      )}
    </>
  );
};

type RepoActionsMenuProps = {
  repoId: string;
};

const RepoActionsMenu: React.FC<RepoActionsMenuProps> = ({repoId}) => {
  const {
    deleteRepoModalOpen,
    setDeleteModalOpen,
    onDropdownMenuSelect,
    menuItems,
    onMenuOpen,
  } = useRepoActionsMenu(repoId);

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
        Repo Actions
      </DefaultDropdown>
      <RepoActionsModals
        repoId={repoId}
        deleteRepoModalOpen={deleteRepoModalOpen}
        setDeleteModalOpen={setDeleteModalOpen}
      />
    </>
  );
};

export default RepoActionsMenu;
