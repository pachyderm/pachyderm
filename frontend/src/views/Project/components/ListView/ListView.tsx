import {ApolloError} from '@apollo/client';
import {Group} from '@pachyderm/components';
import React, {useEffect} from 'react';
import {useHistory} from 'react-router';

import EmptyState from '@dash-frontend/components/EmptyState';
import {
  NO_DAG_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import View from '@dash-frontend/components/View';
import {Node, DagNodes} from '@dash-frontend/lib/types';

import DAGError from '../DAGError';

import ListItem from './components/ListItem';
import styles from './ListView.module.css';

type ListViewProps = {
  items: DagNodes[];
  getNodePath: (node: Node) => string;
  selectedItem?: string;
  error?: ApolloError | string;
};

const ListView: React.FC<ListViewProps> = ({
  items,
  getNodePath,
  selectedItem,
  error,
}) => {
  const browserHistory = useHistory();

  useEffect(() => {
    const nodes = items
      .map((dag) => dag.nodes)
      .flat()
      .filter((node) => node.access);
    if (!selectedItem && nodes.length > 0) {
      browserHistory.push(getNodePath(nodes[0]));
    }
  }, [browserHistory, getNodePath, items, selectedItem]);

  return (
    <View className={styles.view}>
      <Group spacing={16} vertical>
        <DAGError error={error} />
        <div className={styles.base}>
          {items.length > 0 ? (
            items.map(
              (dag) =>
                dag.nodes.length > 0 && (
                  <div className={styles.wrapper} key={dag.id}>
                    {dag.nodes.map((node) => (
                      <ListItem
                        node={node}
                        key={node.id}
                        selectedItem={selectedItem}
                        nodePath={getNodePath(node)}
                      />
                    ))}
                  </div>
                ),
            )
          ) : (
            <EmptyState title={LETS_START_TITLE} message={NO_DAG_MESSAGE} />
          )}
        </div>
      </Group>
    </View>
  );
};

export default ListView;
