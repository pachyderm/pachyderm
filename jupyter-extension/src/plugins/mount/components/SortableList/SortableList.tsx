import React, {useCallback} from 'react';
import {Repo, Mount, ListMountsResponse} from '../../types';
import {caretUpIcon, caretDownIcon} from '@jupyterlab/ui-components';
import {useSort, stringComparator} from '../../../../utils/hooks/useSort';
import ListMount from './ListMount';
import ListUnmount from './ListUnmount';

type SortableListProps = {
  items: Mount[] | Repo[];
  open: (path: string) => void;
  updateData: (data: ListMountsResponse) => void;
  mountedItems: Mount[];
};

const nameComparator = {
  name: 'Name',
  func: stringComparator,
  accessor: (item: Mount | Repo) => ('name' in item ? item.name : item.repo),
};

const SortableList: React.FC<SortableListProps> = ({
  open,
  items,
  updateData,
  mountedItems,
}) => {
  const {sortedData, setComparator, reversed} = useSort<Mount | Repo>({
    data: items,
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
          sortedData.map((item: Mount | Repo) =>
            'name' in item ? (
              <ListMount
                item={item}
                key={item.name}
                open={open}
                updateData={updateData}
              />
            ) : (
              <ListUnmount
                item={item}
                key={item.repo}
                updateData={updateData}
                mountedItems={mountedItems}
              />
            ),
          )}
      </ul>
    </div>
  );
};

export default SortableList;
