import {Circle, CircleColor} from '../../../../utils/components/Circle/Circle';
import {capitalize} from 'lodash';
import React, {useEffect, useState} from 'react';
import {requestAPI} from '../../../../handler';
import {findMountedBranch} from '../../pollRepos';
import {Branch, mountState, Repo} from '../../types';
import {infoIcon} from '../../../../utils/icons';

export const DISABLED_STATES: mountState[] = [
  'unmounting',
  'mounting',
  'error',
];

export const ACCESS: mountState[] = ['unmounting', 'mounting', 'error'];

type ListItemProps = {
  repo: Repo;
  open: (path: string) => void;
  updateData: (data: Repo[]) => void;
};

const ListItem: React.FC<ListItemProps> = ({repo, open, updateData}) => {
  const [mountedBranch, setMountedBranch] = useState<Branch>();
  const [selectedBranch, setSelectedBranch] = useState<string>();
  const [disabled, setDisabled] = useState<boolean>(false);
  const [authorized, setAuthorized] = useState<boolean>(false);

  useEffect(() => {
    setAuthorized(repo.authorization !== 'none');
  }, [repo]);

  useEffect(() => {
    const found = findMountedBranch(repo);
    if (found) {
      setMountedBranch(found);
      setDisabled(DISABLED_STATES.includes(found.mount[0].state));
    }
  }, [repo]);

  useEffect(() => {
    if (repo.branches.length > 0) {
      const found = repo.branches.find((branch) => branch.branch === 'master');
      setSelectedBranch(found ? found.branch : repo.branches[0].branch);
    }
  }, [repo]);

  const onChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedBranch(e.target.value);
  };

  const hasBranches = repo?.branches?.length > 0;
  const buttonText = mountedBranch ? 'Unmount' : 'Mount';

  const openFolder = () => {
    if (mountedBranch) {
      open(repo.repo);
    }
  };

  const onClickHandler = () => {
    mount();
  };

  let behind = -1;
  if (
    typeof mountedBranch !== 'undefined' &&
    mountedBranch.mount.length > 0 &&
    typeof mountedBranch.mount[0].how_many_commits_behind !== 'undefined'
  ) {
    behind = mountedBranch.mount[0].how_many_commits_behind;
  }

  const mount = async () => {
    setDisabled(true);
    try {
      if (mountedBranch) {
        const updatedRepos = await requestAPI<Repo[]>(
          `repos/${repo.repo}/${mountedBranch.branch}/_unmount?name=${repo.repo}`,
          'PUT',
        );
        updateData(updatedRepos);
      } else {
        if (selectedBranch) {
          const updatedRepos = await requestAPI<Repo[]>(
            `repos/${repo.repo}/${selectedBranch}/_mount?name=${repo.repo}&mode=ro`,
            'PUT',
          );
          updateData(updatedRepos);
        }
      }
      open('');
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
          <span className="pachyderm-mount-list-item-name" title={repo.repo}>
            {repo.repo}
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
        title="A repository must have a branch in order to mount it"
      >
        <span className="pachyderm-mount-list-item-name-branch-wrapper pachyderm-mount-sortableList-disabled">
          <span className="pachyderm-mount-list-item-name" title={repo.repo}>
            {repo.repo}
          </span>

          <span className="pachyderm-mount-list-item-branch">No Branches</span>
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
        onClick={openFolder}
      >
        <span className="pachyderm-mount-list-item-name" title={repo.repo}>
          {repo.repo}
        </span>
        <span className="pachyderm-mount-list-item-branch">
          {mountedBranch ? (
            <div>
              <span title={mountedBranch.branch}>@ {mountedBranch.branch}</span>
              <span
                style={{marginLeft: '7px'}}
                data-testid="ListItem__commitBehindness"
              >
                {renderCommitBehindness(behind)}
              </span>
            </div>
          ) : (
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
                {repo.branches.map((branch) => {
                  return (
                    <option key={branch.branch} value={branch.branch}>
                      {branch.branch}
                    </option>
                  );
                })}
              </select>
            </>
          )}
        </span>
      </span>
      <span className="pachyderm-mount-list-item-action">
        <button
          disabled={disabled}
          onClick={onClickHandler}
          className="pachyderm-button-link"
          data-testid={`ListItem__${buttonText.toLowerCase()}`}
        >
          {buttonText}
        </button>
        {mountedBranch && (
          <span
            className="pachyderm-mount-list-item-status"
            data-testid="ListItem__status"
          >
            {renderStatus(
              mountedBranch.mount[0].state,
              mountedBranch.mount[0].status,
            )}
          </span>
        )}
      </span>
    </li>
  );
};

const renderCommitBehindness = (behind: number) => {
  if (behind === 0) {
    return <span>✅ up to date</span>;
  } else if (behind === 1) {
    return <span>⌛ {behind} commit behind</span>;
  } else {
    return <span>⌛ {behind} commits behind</span>;
  }
};

const renderStatus = (state: mountState, status: string | null) => {
  let color = 'gray';
  let statusMessage = '';

  switch (state) {
    case 'mounted':
      color = 'green';
      break;
    case 'unmounting':
    case 'mounting':
      color = 'yellow';
      break;
    case 'error':
      color = 'red';
      break;
  }

  if (status) {
    statusMessage = `${capitalize(state || 'Unknown')}: ${status}`;
  } else {
    statusMessage = capitalize(state || 'Unknown');
  }

  return (
    <>
      <Circle
        color={color as CircleColor}
        className="pachyderm-mount-list-item-status-circle"
      />

      <div
        data-testid="ListItem__statusIcon"
        className="pachyderm-mount-list-item-status-icon"
        title={statusMessage}
      >
        <infoIcon.react tag="span" />
      </div>
    </>
  );
};

export default ListItem;
