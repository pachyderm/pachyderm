import React, {useEffect, useState} from 'react';
import {requestAPI} from '../../../../handler';
import {ListMountsResponse, mountState, Repo, Mount} from '../../types';

export const DISABLED_STATES: mountState[] = [
  'unmounting',
  'mounting',
  'error',
];

type ListUnmountProps = {
  item: Repo;
  updateData: (data: ListMountsResponse) => void;
  mountedItems: Mount[];
};

const ListUnmount: React.FC<ListUnmountProps> = ({
  item,
  updateData,
  mountedItems,
}) => {
  const [selectedBranch, setSelectedBranch] = useState<string>('');
  const [selectedBranchMounted, setSelectedBranchMounted] =
    useState<boolean>(false);
  const [disabled, setDisabled] = useState<boolean>(false);
  const [authorized, setAuthorized] = useState<boolean>(false);
  const hasBranches = item?.branches?.length > 0;
  const buttonText = 'Mount';

  useEffect(() => {
    setAuthorized(item.authorization !== 'none');
  }, [item]);

  useEffect(() => {
    const branchMounted = mountedItems.find(
      (mount) => mount.repo === item.repo && mount.branch === selectedBranch,
    );
    setSelectedBranchMounted(branchMounted ? true : false);
  }, [mountedItems, selectedBranch]);

  useEffect(() => {
    if (hasBranches) {
      const found = item.branches.find((branch) => branch === 'master');
      setSelectedBranch(found ? found : item.branches[0]);
      setDisabled(false);
    }
  }, [item]);

  const onChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedBranch(e.target.value);
  };

  const mount = async () => {
    setDisabled(true);
    try {
      if (selectedBranch) {
        const data = await requestAPI<ListMountsResponse>('_mount', 'PUT', {
          mounts: [
            {
              name:
                selectedBranch === 'master'
                  ? item.repo
                  : `${item.repo}_${selectedBranch}`,
              repo: item.repo,
              branch: selectedBranch,
              mode: 'ro',
            },
          ],
        });
        updateData(data);
      }
    } catch {
      console.log('error mounting or unmounting repo');
    }
  };

  if (!authorized) {
    return (
      <li
        className="pachyderm-mount-sortableList-item"
        data-testid="ListItem__unauthorized"
        style={{cursor: 'not-allowed'}}
        title="You don't have the correct permissions to access this repository"
      >
        <span className="pachyderm-mount-list-item-name-branch-wrapper pachyderm-mount-sortableList-disabled">
          <span className="pachyderm-mount-list-item-name" title={item.repo}>
            {item.repo}
          </span>
          <span className="pachyderm-mount-list-item-branch">
            No read access
          </span>
        </span>
      </li>
    );
  }

  if (!hasBranches) {
    return (
      <li
        className="pachyderm-mount-sortableList-item"
        data-testid="ListItem__noBranches"
        title="Repo doesn't have a branch"
      >
        <span className="pachyderm-mount-list-item-name-branch-wrapper pachyderm-mount-sortableList-disabled">
          <span className="pachyderm-mount-list-item-name" title={item.repo}>
            {item.repo}
          </span>

          <span className="pachyderm-mount-list-item-branch">No branches</span>
        </span>
      </li>
    );
  }

  return (
    <li
      className="pachyderm-mount-sortableList-item"
      data-testid="ListItem__branches"
    >
      <span
        className={`pachyderm-mount-list-item-name-branch-wrapper ${
          disabled ? 'pachyderm-mount-sortableList-disabled' : ''
        }`}
      >
        <span className="pachyderm-mount-list-item-name" title={item.repo}>
          {item.repo}
        </span>
        <span className="pachyderm-mount-list-item-branch">
          {
            <>
              <span>@ </span>
              <select
                disabled={disabled}
                name="branch"
                value={selectedBranch}
                className="pachyderm-mount-list-item-branch-select"
                onChange={onChange}
                data-testid="ListItem__select"
              >
                {item.branches.map((branch) => {
                  return (
                    <option key={branch} value={branch}>
                      {branch}
                    </option>
                  );
                })}
              </select>
            </>
          }
        </span>
      </span>
      <span className="pachyderm-mount-list-item-action">
        <button
          disabled={disabled || selectedBranchMounted}
          onClick={mount}
          className="pachyderm-button-link"
          data-testid={`ListItem__${buttonText.toLowerCase()}`}
        >
          {buttonText}
        </button>
      </span>
    </li>
  );
};

export default ListUnmount;
