/* eslint-disable testing-library/prefer-screen-queries */
import {render, act} from '@testing-library/react';
import React, {useEffect} from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import useLocalProjectSettings from '../useLocalProjectSettings';

const SettingsComponent = withContextProviders(
  ({onRender}: {onRender: jest.Mock}) => {
    const [viewSetting, setViewSetting] = useLocalProjectSettings({
      projectId: 'default',
      key: 'view_setting',
    });
    const [lastAction, setLastAction] = useLocalProjectSettings({
      projectId: 'default',
      key: 'last_action',
    });

    useEffect(() => {
      setLastAction('recording');
      setViewSetting('listView');
    }, [setLastAction, setViewSetting]);

    onRender();

    return (
      <div>
        <div>
          <span>view_setting: </span>
          <span>{viewSetting}</span>
        </div>
        <div>
          <span>last_action: </span>
          <span>{lastAction}</span>
        </div>
        <button onClick={() => setViewSetting('iconView')}>change view</button>
        <button onClick={() => setLastAction('stalling')}>change action</button>
      </div>
    );
  },
);

const ViewComponent = withContextProviders(
  ({onRender}: {onRender: jest.Mock}) => {
    const [viewSetting] = useLocalProjectSettings({
      projectId: 'default',
      key: 'view_setting',
    });

    onRender();

    return (
      <div>
        <span>view-{viewSetting}</span>
      </div>
    );
  },
);

describe('useLocalProjectSettings', () => {
  it('should set and get local storage values', async () => {
    const onRender = jest.fn();
    const onRenderView = jest.fn();

    const {getByText} = render(<SettingsComponent onRender={onRender} />);
    const {getByText: getByTextView} = render(
      <ViewComponent onRender={onRenderView} />,
    );

    expect(getByTextView('view-listView')).toBeInTheDocument();
    expect(getByText('listView')).toBeInTheDocument();
    expect(getByText('recording')).toBeInTheDocument();
    expect(onRender).toHaveBeenCalledTimes(2);
    expect(onRenderView).toHaveBeenCalledTimes(1);

    act(() => getByText('change view').click());

    expect(getByTextView('view-iconView')).toBeInTheDocument();
    expect(getByText('iconView')).toBeInTheDocument();
    expect(getByText('recording')).toBeInTheDocument();

    act(() => getByText('change action').click());

    expect(getByTextView('view-iconView')).toBeInTheDocument();
    expect(getByText('iconView')).toBeInTheDocument();
    expect(getByText('stalling')).toBeInTheDocument();

    expect(onRender).toHaveBeenCalledTimes(4);
    expect(onRenderView).toHaveBeenCalledTimes(2);
  });
});
