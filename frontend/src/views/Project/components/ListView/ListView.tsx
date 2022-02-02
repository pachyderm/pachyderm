import React, {useEffect} from 'react';
import {useHistory} from 'react-router';

import EmptyState from '@dash-frontend/components/EmptyState';
import {
  NO_DAG_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import View from '@dash-frontend/components/View';
import {Node, DagNodes} from '@dash-frontend/lib/types';
import {useWorkspace} from 'hooks/useWorkspace';

import ListItem from './components/ListItem';
import styles from './ListView.module.css';

type ListViewProps = {
  items: DagNodes[];
  getNodePath: (node: Node) => string;
  selectedItem?: string;
};

const ListView: React.FC<ListViewProps> = ({
  items,
  getNodePath,
  selectedItem,
}) => {
  const browserHistory = useHistory();
  const {hasConnectInfo} = useWorkspace();

  useEffect(() => {
    if (!selectedItem && items.length > 0 && items[0].nodes.length > 0) {
      browserHistory.push(getNodePath(items[0].nodes[0]));
    }
  }, [browserHistory, getNodePath, items, selectedItem]);

  return (
    <View className={styles.view}>
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
          <EmptyState
            title={LETS_START_TITLE}
            message={NO_DAG_MESSAGE}
            connect={hasConnectInfo}
          />
        )}
      </div>
    </View>
  );
};

export default ListView;
