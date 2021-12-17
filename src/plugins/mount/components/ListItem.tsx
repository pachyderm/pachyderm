import React, {useEffect, useState} from 'react';
import {requestAPI} from '../../../handler';
import {Repo} from '../mount';

type ListItemProps = {
  repo: Repo;
  open: (path: string) => void;
};

const ListItem: React.FC<ListItemProps> = ({repo, open}) => {
  const [mountedBanch, setMountedBranch] = useState<string>();
  const [selectedBranch, setSelectedBranch] = useState<string>();

  useEffect(() => {
    const found = repo.branches.find(
      (branch) => branch.mount.state === 'mounted',
    );
    if (found) {
      setMountedBranch(found.branch);
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
  const buttonText = mountedBanch ? 'Unmount' : 'Mount';

  const openFolder = () => {
    if (mountedBanch) {
      open(repo.repo);
    }
  };

  const onClickHandler = (
    e: React.MouseEvent<HTMLAnchorElement, MouseEvent>,
  ) => {
    e.stopPropagation();
    mount();
  };

  const mount = async () => {
    if (mountedBanch) {
      await requestAPI<any>(
        `repos/${repo.repo}/${mountedBanch}/_unmount?name=${repo.repo}`,
        'PUT',
      );
    } else {
      await requestAPI<any>(
        `repos/${repo.repo}/${selectedBranch}/_mount?name=${repo.repo}&mode=ro`,
        'PUT',
      );
    }
    open('');
  };

  if (!hasBranches) {
    return (
      <li className="pachyderm-mount-sortableList-item">
        <span className="pachyderm-mount-list-item-name pachyderm-mount-sortableList-item-no-branchs">
          {repo.repo}
        </span>
        <span className="pachyderm-mount-list-item-name pachyderm-mount-sortableList-item-no-branchs">
          No Branches
        </span>
      </li>
    );
  }
  return (
    <li className="pachyderm-mount-sortableList-item" onClick={openFolder}>
      <span className="pachyderm-mount-list-item-name">{repo.repo}</span>
      <span className="pachyderm-mount-list-item-branch">
        {mountedBanch ? (
          <>{mountedBanch}</>
        ) : (
          <select
            name="branch"
            value={selectedBranch}
            className="pachyderm-mount-list-item-branch-select"
            onChange={onChange}
          >
            {repo.branches.map((branch) => {
              return (
                <option key={branch.branch} value={branch.branch}>
                  {branch.branch}
                </option>
              );
            })}
          </select>
        )}
      </span>
      <span className="pachyderm-mount-list-item-action">
        <a onClick={onClickHandler} className="pachyderm-help-link">
          {buttonText}
        </a>
      </span>
    </li>
  );
};

export default ListItem;
