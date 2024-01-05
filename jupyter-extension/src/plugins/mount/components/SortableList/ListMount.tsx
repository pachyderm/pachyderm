import React, {useState} from 'react';
import {requestAPI} from '../../../../handler';
import {Mount, ListMountsResponse} from '../../types';

type ListMountProps = {
  item: Mount;
  open: (path: string) => void;
  updateData: (data: ListMountsResponse) => void;
};

const ListMount: React.FC<ListMountProps> = ({item, open, updateData}) => {
  const [disabled, setDisabled] = useState<boolean>(false);
  const branch = item.branch;
  const buttonText = 'Unload';

  const openFolder = () => {
    open(item.name);
  };

  const unmount = async () => {
    setDisabled(true);
    try {
      const data = await requestAPI<ListMountsResponse>('_unmount', 'PUT', {
        mounts: [`${item.name}`],
      });
      updateData(data);
    } catch {
      console.log('error unmounting repo');
    }
  };

  return (
    <li
      className="pachyderm-mount-sortableList-item"
      data-testid="ListItem__repo"
    >
      <span
        className={`pachyderm-mount-list-item-name-branch-wrapper ${
          disabled ? 'pachyderm-mount-sortableList-disabled' : ''
        }`}
        onClick={openFolder}
      >
        <span className="pachyderm-mount-list-item-name" title={item.name}>
          {item.name}
        </span>
        <span className="pachyderm-mount-list-item-branch">
          {
            <div>
              <span title={branch}>@ {branch}</span>
            </div>
          }
        </span>
      </span>
      <span className="pachyderm-mount-list-item-action">
        <button
          disabled={disabled}
          onClick={unmount}
          className="pachyderm-button-link"
          data-testid={`ListItem__${buttonText.toLowerCase()}`}
        >
          {buttonText}
        </button>
      </span>
    </li>
  );
};

export default ListMount;
