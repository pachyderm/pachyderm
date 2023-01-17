import {render, waitFor, screen} from '@testing-library/react';
import React from 'react';
import xhrMock from 'xhr-mock';

import {Experiment, Variation, initFeatureFlagsProvider} from '../';

const FeatureFlagsProvider = initFeatureFlagsProvider('12345678');

describe('FeatureFlags', () => {
  beforeEach(() => {
    xhrMock.setup();
    xhrMock.post(/.*/, {
      status: 200,
      body: '',
    });
  });

  afterEach(() => {
    xhrMock.teardown();
  });

  it('should show a component with a flag on', async () => {
    xhrMock.get(/.*/, {
      body: '{"test-flag":{"flagVersion":5,"trackEvents":false,"value":true,"variation":0,"version":5}}',
      headers: {
        'content-type': 'application/json',
      },
      status: 200,
    });

    render(
      <FeatureFlagsProvider>
        <Experiment name="testFlag">
          <span>Static Component</span>
          <Variation value={true}>Test Flag On</Variation>
          <Variation value={false}>Test Flag Off</Variation>
        </Experiment>
      </FeatureFlagsProvider>,
    );

    expect(screen.getByText('Static Component')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('Test Flag On')).toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.queryByText('Test Flag Off')).not.toBeInTheDocument();
    });
  });

  it('should show a component with a flag off', async () => {
    xhrMock.get(/.*/, {
      body: '{"test-flag":{"flagVersion":5,"trackEvents":false,"value":false,"variation":0,"version":5}}',
      headers: {
        'content-type': 'application/json',
      },
      status: 200,
    });

    render(
      <FeatureFlagsProvider>
        <Experiment name="testFlag">
          <span>Static Component</span>
          <Variation value={true}>Test Flag On</Variation>
          <Variation value={false}>Test Flag Off</Variation>
        </Experiment>
      </FeatureFlagsProvider>,
    );

    expect(screen.getByText('Static Component')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.queryByText('Test Flag On')).not.toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.getByText('Test Flag Off')).toBeInTheDocument();
    });
  });

  it('should not render either variation when a flag does not exist', async () => {
    xhrMock.get(/.*/, {
      body: '{"test-flag":{"flagVersion":5,"trackEvents":false,"value":true,"variation":0,"version":5}}',
      headers: {
        'content-type': 'application/json',
      },
      status: 200,
    });

    render(
      <FeatureFlagsProvider>
        <Experiment name="testFlag">
          <Variation value={true}>Test Flag On</Variation>
          <Variation value={false}>Test Flag Off</Variation>
        </Experiment>
        <Experiment name="testBadFlag">
          <Variation value={true}>Test Bad Flag On</Variation>
          <Variation value={false}>Test Bad Flag Off</Variation>
        </Experiment>
      </FeatureFlagsProvider>,
    );

    await waitFor(() => {
      expect(screen.getByText('Test Flag On')).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.queryByText('Test Flag Off')).not.toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.queryByText('Test Bad Flag On')).not.toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.queryByText('Test Bad Flag Off')).not.toBeInTheDocument();
    });
  });
});
