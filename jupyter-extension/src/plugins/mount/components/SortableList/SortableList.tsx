import React, {useCallback} from 'react';
import {Repo} from '../../types';
import ListItem from './ListItem';
import {caretUpIcon, caretDownIcon} from '@jupyterlab/ui-components';
import {useSort, stringComparator} from '../../../../utils/hooks/useSort';

type SortableListProps = {
  repos: Repo[];
  open: (path: string) => void;
  updateData: (data: Repo[]) => void;
};

const nameComparator = {
  name: 'Name',
  func: stringComparator,
  accessor: (repo: Repo) => repo.repo,
};

const SortableList: React.FC<SortableListProps> = ({
  open,
  repos,
  updateData,
}) => {
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
      </div>
      <ul className="pachyderm-mount-sortableList-content">
        {sortedData &&
          sortedData.map((repo: Repo) => (
            <ListItem
              repo={repo}
              key={repo.repo}
              open={open}
              updateData={updateData}
            />
          ))}
      </ul>
    </div>
  );
};

export default SortableList;
