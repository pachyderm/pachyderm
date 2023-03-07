import {NodeState} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import useUrlQueryState from '../useUrlQueryState';

const ViewStateComponent = withContextProviders(({onRender}) => {
  const {
    viewState,
    getNewViewState,
    updateViewState,
    clearViewState,
    toggleSelection,
  } = useUrlQueryState();

  onRender();

  return (
    <div>
      {Object.entries(viewState).length === 0 && <span>Empty State</span>}
      {Object.entries(viewState).map(([key, value]) => (
        <span key={key}>{`${key}: ${value}`}</span>
      ))}
      <button
        onClick={() =>
          getNewViewState({sortBy: 'Created: Newest', globalIdFilter: ''})
        }
      >
        getNewViewState
      </button>
      <button
        onClick={() =>
          updateViewState({jobStatus: [NodeState.PAUSED], jobId: []})
        }
      >
        updateViewState
      </button>
      <button onClick={() => clearViewState()}>clearViewState</button>
      <button onClick={() => toggleSelection('selectedPipelines', 'edges')}>
        toggleSelectionEdges
      </button>
      <button onClick={() => toggleSelection('selectedPipelines', 'montage')}>
        toggleSelectionMontage
      </button>
    </div>
  );
});

describe('useUrlQueryState', () => {
  it('should create, update and clear viewstate', async () => {
    const onRender = jest.fn();
    render(<ViewStateComponent onRender={onRender} />);

    expect(onRender).toHaveBeenCalledTimes(1);

    screen.getByText('getNewViewState').click();
    expect(screen.getByText('sortBy: Created: Newest')).toBeInTheDocument();

    screen.getByText('updateViewState').click();
    expect(screen.getByText('sortBy: Created: Newest')).toBeInTheDocument();
    expect(screen.getByText('jobStatus: PAUSED')).toBeInTheDocument();

    screen.getByText('clearViewState').click();
    expect(screen.getByText('Empty State')).toBeInTheDocument();

    expect(onRender).toHaveBeenCalledTimes(4);
  });

  it('should toggle list selections', async () => {
    const onRender = jest.fn();
    render(<ViewStateComponent onRender={onRender} />);

    expect(onRender).toHaveBeenCalledTimes(1);

    screen.getByText('toggleSelectionEdges').click();
    expect(screen.getByText('selectedPipelines: edges')).toBeInTheDocument();
    screen.getByText('toggleSelectionMontage').click();
    expect(
      screen.getByText('selectedPipelines: edges,montage'),
    ).toBeInTheDocument();
    screen.getByText('toggleSelectionEdges').click();
    expect(screen.getByText('selectedPipelines: montage')).toBeInTheDocument();

    expect(onRender).toHaveBeenCalledTimes(4);
  });
});
