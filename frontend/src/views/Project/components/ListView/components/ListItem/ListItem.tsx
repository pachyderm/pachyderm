import {NodeState, NodeType} from '@graphqlTypes';
import {
  Link,
  DocumentAddSVG,
  Icon,
  StatusBlockedSVG,
  StatusPausedSVG,
  CaptionTextSmall,
} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';

import deriveRepoNameFromNode from '@dash-frontend/lib/deriveRepoNameFromNode';
import {Node} from '@dash-frontend/lib/types';

import styles from './ListItem.module.css';

type ListItemProps = {
  node: Node;
  selectedItem?: string;
  nodePath: string;
};

const getIcon = (node: Node) => {
  switch (node.type) {
    case NodeType.INPUT_REPO:
    case NodeType.OUTPUT_REPO:
      return '/dag_repo.svg';
    case NodeType.PIPELINE:
      return '/dag_pipeline.svg';
    default:
      return undefined;
  }
};

const ListItem: React.FC<ListItemProps> = ({node, selectedItem, nodePath}) => {
  const nodeName =
    node.type !== NodeType.PIPELINE ? deriveRepoNameFromNode(node) : node.id;

  return (
    <Link
      data-testid="ListItem__row"
      aria-label={node.id}
      to={node.access ? nodePath : undefined}
      className={classnames(styles.base, {
        [styles[node.type]]: true,
        [styles.selected]: selectedItem === nodeName,
        [styles.error]: node.state === NodeState.ERROR,
        [styles.paused]: node.state === NodeState.PAUSED,
      })}
    >
      <img className={styles.nodeImage} src={getIcon(node)} alt={node.type} />
      <span className={styles.nodeName}>{nodeName}</span>
      {NodeType.INPUT_REPO === node.type && (
        <CaptionTextSmall className={styles.status}>
          Input{' '}
          <Icon small>
            <DocumentAddSVG />
          </Icon>
        </CaptionTextSmall>
      )}
      {NodeType.PIPELINE === node.type && node.state === NodeState.ERROR && (
        <Icon small>
          <StatusBlockedSVG />
        </Icon>
      )}
      {NodeType.PIPELINE === node.type && node.state === NodeState.PAUSED && (
        <Icon small>
          <StatusPausedSVG />
        </Icon>
      )}
    </Link>
  );
};

export default ListItem;
