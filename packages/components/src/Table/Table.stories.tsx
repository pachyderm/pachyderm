import React from 'react';

import Table from './';

export default {title: 'Table'};

const bruinsPointLeaders = [
  {id: 0, name: 'Brad Marchand', points: 7},
  {id: 1, name: 'David Krejci', points: 7},
  {id: 2, name: 'Patrice Bergeron', points: 4},
  {id: 3, name: 'Charlie McAvoy', points: 4},
  {id: 4, name: 'Ondrej Kase', points: 4},
  {id: 5, name: 'Jake Debrusk', points: 3},
  {id: 6, name: 'Charlie Coyle', points: 3},
];

export const Default = () => {
  return (
    <Table>
      <Table.Head>
        <Table.HeaderCell>Name</Table.HeaderCell>
        <Table.HeaderCell>Points</Table.HeaderCell>
      </Table.Head>

      <Table.Body bandedRows>
        {bruinsPointLeaders.map((player) => (
          <Table.Row key={player.id}>
            <Table.DataCell>{player.name}</Table.DataCell>
            <Table.DataCell>{player.points}</Table.DataCell>
          </Table.Row>
        ))}
      </Table.Body>
    </Table>
  );
};

export const StickyHeader = () => {
  return (
    <div style={{maxHeight: 330, overflow: 'scroll'}}>
      <Table>
        <Table.Head>
          <Table.Row sticky>
            <Table.HeaderCell>Name</Table.HeaderCell>
            <Table.HeaderCell>Points</Table.HeaderCell>
          </Table.Row>
        </Table.Head>

        <Table.Body bandedRows>
          {bruinsPointLeaders.map((player) => (
            <Table.Row key={player.id}>
              <Table.DataCell>{player.name}</Table.DataCell>
              <Table.DataCell>{player.points}</Table.DataCell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    </div>
  );
};
