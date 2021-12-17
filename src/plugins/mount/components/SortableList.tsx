import React, {useCallback} from 'react';
import {Repo} from '../mount';
import ListItem from './ListItem';
import {caretUpIcon, caretDownIcon} from '@jupyterlab/ui-components';
import {useSort, stringComparator} from '@pachyderm/components';

type SortableListProps = {
  repos: Repo[];
  open: (path: string) => void;
};

const nameComparator = {
  name: 'Name',
  func: stringComparator,
  accessor: (repo: Repo) => repo.repo,
};

const SortableList: React.FC<SortableListProps> = ({open, repos}) => {
  const {sortedData, setComparator, reversed} = useSort<Repo>({
    data: repos,
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
