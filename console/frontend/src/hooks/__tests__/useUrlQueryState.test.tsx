import {render, screen, act} from '@testing-library/react';
import React from 'react';

import {NodeState} from '@dash-frontend/lib/types';
import {withContextProviders} from '@dash-frontend/testHelpers';

import useUrlQueryState from '../useUrlQueryState';

const ViewStateComponent = withContextProviders(
  ({onRender}: {onRender: jest.Mock}) => {
    const {
      searchParams,
      getNewSearchParamsAndGo,
      updateSearchParamsAndGo,
      clearSearchParamsAndGo,
      toggleSearchParamsListEntry,
    } = useUrlQueryState();

    onRender();

    return (
      <div>
        {Object.entries(searchParams).length === 0 && <span>Empty State</span>}
        {Object.entries(searchParams).map(([key, value]) => (
          <span key={key}>{`${key}: ${value}`}</span>
        ))}
        <button
          onClick={() =>
            getNewSearchParamsAndGo({
              sortBy: 'Created: Newest',
              globalIdFilter: '',
            })
          }
        >
          getNewSearchParamsAndGo
        </button>
        <button
          onClick={() =>
            updateSearchParamsAndGo({jobStatus: [NodeState.PAUSED], jobId: []})
          }
        >
          updateSearchParamsAndGo
        </button>
        <button onClick={() => clearSearchParamsAndGo()}>
          clearSearchParamsAndGo
        </button>
        <button
          onClick={() =>
            toggleSearchParamsListEntry('selectedPipelines', 'edges')
          }
        >
          toggleSelectionEdges
        </button>
        <button
          onClick={() =>
            toggleSearchParamsListEntry('selectedPipelines', 'montage')
          }
        >
          toggleSelectionMontage
        </button>
      </div>
    );
  },
);

describe('useUrlQueryState', () => {
  it('should create, update and clear viewstate', async () => {
    const onRender = jest.fn();
    render(<ViewStateComponent onRender={onRender} />);

    expect(onRender).toHaveBeenCalledTimes(1);

    act(() => screen.getByText('getNewSearchParamsAndGo').click());
    expect(screen.getByText('sortBy: Created: Newest')).toBeInTheDocument();

    act(() => screen.getByText('updateSearchParamsAndGo').click());
    expect(screen.getByText('sortBy: Created: Newest')).toBeInTheDocument();
    expect(screen.getByText('jobStatus: PAUSED')).toBeInTheDocument();

    act(() => screen.getByText('clearSearchParamsAndGo').click());
    expect(screen.getByText('Empty State')).toBeInTheDocument();

    expect(onRender).toHaveBeenCalledTimes(4);
  });

  it('should toggle list selections', async () => {
    const onRender = jest.fn();
    render(<ViewStateComponent onRender={onRender} />);

    expect(onRender).toHaveBeenCalledTimes(1);

    act(() => screen.getByText('toggleSelectionEdges').click());
    expect(screen.getByText('selectedPipelines: edges')).toBeInTheDocument();
    act(() => screen.getByText('toggleSelectionMontage').click());
    expect(
      screen.getByText('selectedPipelines: edges,montage'),
    ).toBeInTheDocument();
    act(() => screen.getByText('toggleSelectionEdges').click());
    expect(screen.getByText('selectedPipelines: montage')).toBeInTheDocument();

    expect(onRender).toHaveBeenCalledTimes(4);
  });
});
