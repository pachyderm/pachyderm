import React, {useCallback, useEffect, useRef, useState} from 'react';
import {Repo} from '../mount';
import ListItem from './ListItem';
import {caretUpIcon, caretDownIcon} from '@jupyterlab/ui-components';
import {useSort, stringComparator} from '@pachyderm/components';
import {requestAPI} from '../../../handler';
import partition from 'lodash/partition';

type SortableListProps = {
  open: (path: string) => void;
  type: 'mounted' | 'unmounted';
};

const nameComparator = {
  name: 'Name',
  func: stringComparator,
  accessor: (repo: Repo) => repo.repo,
};

const SortableList: React.FC<SortableListProps> = ({open, type}) => {
  const [repoData, setRepoData] = useState<Repo[]>([]);
  const dataRef = useRef<Repo[]>([]);

  const getRepos = async () => {
    try {
      const data = await requestAPI<Repo[]>('repos', 'GET');

      if (JSON.stringify(data) !== JSON.stringify(dataRef.current)) {
        dataRef.current = data;
        const [mounted, unmounted] = partition(data, (rep: Repo) =>
          rep.branches.find((branch) => branch.mount.state === 'mounted'),
        );
        if (type === 'mounted') {
          setRepoData(mounted);
        } else {
          setRepoData(unmounted);
        }
      }
    } catch {
      console.log('Error getting repos.');
    }
  };

  useEffect(() => {
    getRepos();
    const id = setInterval(getRepos, 2000);
    return () => clearInterval(id);
  }, []);

  const {sortedData, setComparator, reversed} = useSort<Repo>({
    data: repoData,
    initialSort: nameComparator,
    initialDirection: 1,
  });

  const nameClick = useCallback(() => {
    setComparator(nameComparator);
  }, [setComparator]);

  return (
    <div className="pachyderm-mount-sortableList">
      <div className="pachyderm-mount-sortableList-header">
        <div
          className="pachyderm-mount-sortableList-headerItem"
          onClick={nameClick}
        >
          <span>Name</span>
          <span>
            {reversed ? <caretDownIcon.react /> : <caretUpIcon.react />}
          </span>
        </div>
        <div className="pachyderm-mount-sortableList-headerItem-branch">
          <span>Branch</span>
        </div>
      </div>
      <ul className="pachyderm-mount-sortableList-content">
        {sortedData &&
          sortedData.map((repo: Repo) => (
            <ListItem repo={repo} key={repo.repo} open={open} />
          ))}
      </ul>
    </div>
  );
};

export default SortableList;
