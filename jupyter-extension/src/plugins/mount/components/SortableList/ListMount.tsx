import {Circle, CircleColor} from '../../../../utils/components/Circle/Circle';
import {capitalize} from 'lodash';
import React, {useEffect, useState} from 'react';
import {requestAPI} from '../../../../handler';
import {mountState, Mount, ListMountsResponse} from '../../types';
import {infoIcon} from '../../../../utils/icons';

export const DISABLED_STATES: mountState[] = [
  'unmounting',
  'mounting',
  'error',
];

type ListMountProps = {
  item: Mount;
  open: (path: string) => void;
  updateData: (data: ListMountsResponse) => void;
};

const ListMount: React.FC<ListMountProps> = ({item, open, updateData}) => {
  const [disabled, setDisabled] = useState<boolean>(false);
  const branch = item.branch;
  const buttonText = 'Unmount';
  const behind = item.how_many_commits_behind;

  useEffect(() => {
    setDisabled(DISABLED_STATES.includes(item.state));
  }, [item]);

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
      open('');
    } catch {
      console.log('error unmounting repo');
    }
  };

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
        <span className="pachyderm-mount-list-item-name" title={item.name}>
          {item.name}
        </span>
        <span className="pachyderm-mount-list-item-branch">
          {
            <div>
              <span title={branch}>@ {branch}</span>
              <span
                style={{marginLeft: '7px'}}
                data-testid="ListItem__commitBehindness"
              >
                {renderCommitBehindness(behind)}
              </span>
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
        {
          <span
            className="pachyderm-mount-list-item-status"
            data-testid="ListItem__status"
          >
            {renderStatus(item.state, item.status)}
          </span>
        }
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

export default ListMount;
