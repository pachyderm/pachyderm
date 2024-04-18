import classNames from 'classnames';
import React, {useEffect, useState} from 'react';

import {usePipeline} from '@dash-frontend/hooks/usePipeline';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {Link, Node, NodeType} from '@dash-frontend/lib/types';
import {
  ArrowRightSVG,
  CaptionTextSmall,
  EgressSVG,
  PipelineSVG,
  RepoSVG,
  SkeletonBodyText,
  usePreviousValue,
} from '@pachyderm/components';

import useHoveredNode from '../../providers/HoveredNodeProvider/hooks/useHoveredNode';

import styles from './HoverStats.module.css';

const isRepo = (type: NodeType) =>
  type === NodeType.REPO ||
  type === NodeType.CROSS_PROJECT_REPO ||
  type === NodeType.CRON_REPO;

const LinkStats: React.FC<{link: Link}> = ({link}) => {
  return (
    <div className={styles.base} data-testid="HoverStats__link">
      <div className={styles.border}>
        <h6>
          <RepoSVG /> {link.source.name}
        </h6>
      </div>
      <ArrowRightSVG />
      <div className={styles.border}>
        <h6>
          <PipelineSVG /> {link.target.name}
        </h6>
      </div>
    </div>
  );
};

const NodeStats: React.FC<{node: Node; simplifiedView: boolean}> = ({
  node,
  simplifiedView,
}) => {
  const {searchParams} = useUrlQueryState();
  const globalId = searchParams.globalIdFilter;

  const repo = isRepo(node.type);
  const {loading, pipeline} = usePipeline(
    {
      pipeline: {
        name: node.name,
        project: {name: node.project},
      },
      details: true,
    },
    !repo,
  );
  if (repo) {
    return (
      <div className={styles.base}>
        <h6>
          <RepoSVG /> {node.name}
        </h6>
      </div>
    );
  } else if (node.type === NodeType.EGRESS) {
    return (
      <div className={styles.base}>
        <h6>
          <EgressSVG /> Egress
        </h6>
      </div>
    );
  }

  if (!simplifiedView && globalId) {
    return <></>;
  }
  return (
    <div className={classNames(styles.base, styles.flexColumn)}>
      {simplifiedView && (
        <h6>
          <PipelineSVG /> {node.name}
        </h6>
      )}

      {!globalId && (
        <div className={styles.pipelineDetails}>
          <CaptionTextSmall>Workers active</CaptionTextSmall>
          {!loading ? (
            <>
              {pipeline?.details?.workersAvailable}/
              {pipeline?.details?.workersRequested}
            </>
          ) : (
            <SkeletonBodyText />
          )}
        </div>
      )}
    </div>
  );
};

export type HoverStatsProps = {
  nodes?: Node[];
  links?: Link[];
  simplifiedView: boolean;
};
const HoverStats: React.FC<HoverStatsProps> = ({
  nodes,
  links,
  simplifiedView,
}) => {
  const [debouncedValue, setDebouncedValue] = useState('');
  const {hoveredNode} = useHoveredNode();
  const prevHoveredNode = usePreviousValue(hoveredNode);

  useEffect(() => {
    if (hoveredNode === '') {
      setDebouncedValue('');
    } else {
      const handler = setTimeout(() => {
        setDebouncedValue(hoveredNode);
      }, 700);
      return () => {
        clearTimeout(handler);
      };
    }
  }, [hoveredNode, prevHoveredNode]);

  if (!debouncedValue || prevHoveredNode !== hoveredNode) {
    return <></>;
  }

  const foundLink =
    !simplifiedView &&
    links &&
    links.length !== 0 &&
    links.find((link) => link.id === debouncedValue);

  const foundNode =
    nodes &&
    nodes.length !== 0 &&
    nodes.find(
      (node) =>
        node.id === debouncedValue &&
        (simplifiedView ||
          (!isRepo(node.type) && !(node.type === NodeType.EGRESS))),
    );

  return (
    <div>
      {foundLink && <LinkStats key={foundLink.id} link={foundLink} />}
      {foundNode && (
        <NodeStats
          key={foundNode.id}
          node={foundNode}
          simplifiedView={simplifiedView}
        />
      )}
    </div>
  );
};

export default HoverStats;
