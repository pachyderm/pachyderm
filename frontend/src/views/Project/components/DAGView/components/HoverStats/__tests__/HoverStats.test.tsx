import {render, screen} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React, {useEffect} from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {ListPipelineRequest, PipelineInfo} from '@dash-frontend/api/pps';
import {NodeType} from '@dash-frontend/lib/types';
import {buildPipeline} from '@dash-frontend/mocks';
import {withContextProviders} from '@dash-frontend/testHelpers';

import HoveredNodeProvider from '../../../providers/HoveredNodeProvider';
import useHoveredNode from '../../../providers/HoveredNodeProvider/hooks/useHoveredNode';
import {default as HoverStatsComponent} from '../HoverStats';

const SetHoverNode: React.FC<{
  hoveredNode?: string;
}> = ({hoveredNode}) => {
  const {setHoveredNode} = useHoveredNode();
  useEffect(() => setHoveredNode(hoveredNode), [hoveredNode, setHoveredNode]);
  return <></>;
};

const HoverStats = withContextProviders(({props, hoveredNode}) => {
  return (
    <HoveredNodeProvider>
      <HoverStatsComponent {...props} />
      <SetHoverNode hoveredNode={hoveredNode} />
    </HoveredNodeProvider>
  );
});

describe('HoverStats', () => {
  const server = setupServer();

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    server.resetHandlers();
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo>(
        '/api/pps_v2.API/InspectPipeline',
        async (_req, res, ctx) => {
          return res(
            ctx.json(
              buildPipeline({
                pipeline: {name: 'montage'},
                details: {
                  workersAvailable: '2',
                  workersRequested: '12',
                },
              }),
            ),
          );
        },
      ),
    );
  });

  afterAll(() => server.close());

  describe('HoverStats', () => {
    it('should hide hover info for repo in details view and show it in simple view', async () => {
      const nodes = [
        {
          id: '8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65',
          project: 'default',
          name: 'images',
          type: NodeType.REPO,
        },
      ];

      const {rerender} = render(
        <HoverStats
          props={{nodes, simplifiedView: false}}
          hoveredNode={'8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65'}
        />,
      );
      expect(screen.queryByText('images')).not.toBeInTheDocument();

      rerender(
        <HoverStats
          props={{nodes, simplifiedView: true}}
          hoveredNode={'8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65'}
        />,
      );

      await screen.findByText('images');
    });

    it('should hide pipeline name in details view and show pipeline name and worker stats in simple view', async () => {
      const nodes = [
        {
          id: '52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6',
          project: 'default',
          name: 'montage',
          type: NodeType.PIPELINE,
        },
      ];

      const {rerender} = render(
        <HoverStats
          props={{nodes, simplifiedView: false}}
          hoveredNode={'52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6'}
        />,
      );
      expect(screen.queryByText('montage')).not.toBeInTheDocument();
      await screen.findByText('Workers active');
      await screen.findByText('2/12');

      rerender(
        <HoverStats
          props={{nodes, simplifiedView: true}}
          hoveredNode={'52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6'}
        />,
      );

      await screen.findByText('montage');
      await screen.findByText('Workers active');
      await screen.findByText('2/12');
    });

    it('should show edge info in details view and hide info in simple view', async () => {
      const links = [
        {
          id: '90836a7019de3585928b2d2ff840397d80b3d316',

          source: {
            id: '8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65',
            name: 'images',
          },
          target: {
            id: '52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6',
            name: 'montage',
          },
        },
      ];

      const {rerender} = render(
        <HoverStats
          props={{links, simplifiedView: false}}
          hoveredNode={'90836a7019de3585928b2d2ff840397d80b3d316'}
        />,
      );

      await screen.findByText('images');
      await screen.findByText('montage');

      rerender(
        <HoverStats
          props={{links, simplifiedView: true}}
          hoveredNode={'90836a7019de3585928b2d2ff840397d80b3d316'}
        />,
      );

      expect(screen.queryByText('images')).not.toBeInTheDocument();
      expect(screen.queryByText('montage')).not.toBeInTheDocument();
    });

    it('should hide hover info for egress in details view and show it in simple view', async () => {
      const nodes = [
        {
          id: '6857fffb2df6ad89eec23c6717686069659b4420',
          project: 'default',
          name: 's3://test',
          type: NodeType.EGRESS,
        },
      ];

      const {rerender} = render(
        <HoverStats
          props={{nodes, simplifiedView: false}}
          hoveredNode={'6857fffb2df6ad89eec23c6717686069659b4420'}
        />,
      );
      expect(screen.queryByText('Egress')).not.toBeInTheDocument();

      rerender(
        <HoverStats
          props={{nodes, simplifiedView: true}}
          hoveredNode={'6857fffb2df6ad89eec23c6717686069659b4420'}
        />,
      );

      await screen.findByText('Egress');
    });

    it('should hide pipeline info for global id details view and show pipeline name in simple view', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default?globalIdFilter=ef4a6cf0a58641bc81afed33a9adcfa7',
      );

      const nodes = [
        {
          id: '52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6',
          project: 'default',
          name: 'montage',
          type: NodeType.PIPELINE,
        },
      ];

      const {rerender} = render(
        <HoverStats
          props={{nodes, simplifiedView: false}}
          hoveredNode={'52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6'}
        />,
      );
      expect(screen.queryByText('montage')).not.toBeInTheDocument();
      expect(screen.queryByText('Workers active')).not.toBeInTheDocument();

      rerender(
        <HoverStats
          props={{nodes, simplifiedView: true}}
          hoveredNode={'52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6'}
        />,
      );

      await screen.findByText('montage');
      expect(screen.queryByText('Workers active')).not.toBeInTheDocument();
    });
  });
});
