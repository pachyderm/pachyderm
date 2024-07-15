import React from 'react';

import {
  DefaultDropdown,
  ChevronDownSVG,
  OverflowSVG,
  ButtonTypeOptions,
} from '@pachyderm/components';

import DeleteRepoModal from './components/DeleteRepoModal';
import useRepoActionsMenu from './hooks/useRepoActionsMenu';
import styles from './RepoActionsMenu.module.css';

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
        Repo Actions
      </DefaultDropdown>
      <DefaultDropdown
        className={styles.responsiveContent}
        buttonOpts={{
          IconSVG: OverflowSVG,
          ...buttonOpts,
        }}
        aria-label="repo-actions-menu"
        {...dropdownClasses}
      />
      <RepoActionsModals
        repoId={repoId}
        deleteRepoModalOpen={deleteRepoModalOpen}
        setDeleteModalOpen={setDeleteModalOpen}
      />
    </div>
  );
};

export default RepoActionsMenu;
