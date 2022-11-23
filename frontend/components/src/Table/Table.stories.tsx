import React, {useState} from 'react';

import {useSort, stringComparator, numberComparator} from '../hooks/useSort';

import Table from './Table';

export default {title: 'Table'};

type Player = {
  id: number;
  name: string;
  points: number;
};

const bruinsPointLeaders = [
  {id: 0, name: 'Brad Marchand', points: 8},
  {id: 1, name: 'David Krejci', points: 7},
  {id: 2, name: 'Patrice Bergeron', points: 7},
  {id: 3, name: 'Charlie McAvoy', points: 6},
  {id: 4, name: 'Ondrej Kase', points: 5},
  {id: 5, name: 'Jake Debrusk', points: 3},
  {id: 6, name: 'Charlie Coyle', points: 3},
  {id: 7, name: 'Trent Frederic', points: 4},
  {id: 8, name: 'Jack Studnika', points: 4},
  {id: 9, name: 'Jakub Zboril', points: 2},
  {id: 10, name: 'Matt Grzelcyk', points: 3},
  {id: 11, name: 'Hampus Lindholm', points: 1},
];

const datumView = [
  {
    date: 'Sep 30, 2022; 20:32pm',
    jobs: '9 Jobs Total',
    runtime: '4 hours',
    id: 'd386b7b8a79e43e8a3a22ef754bcdc90',
  },
  {
    date: 'Sep 30, 2022; 20:32pm',
    jobs: '9 Jobs Total',
    runtime: '4 hours',
    id: '9a961a7b812e4f5293be02a5b3dace6b',
  },
  {
    date: 'Sep 30, 2022; 20:32pm',
    jobs: '9 Jobs Total',
    runtime: '4 hours',
    id: '8659085713e94556bb02f113d7d789e6',
  },
  {
    date: 'Sep 30, 2022; 20:32pm',
    jobs: '9 Jobs Total',
    runtime: '4 hours',
    id: 'bc7a1ac3df5a450ab0f06117a11d3095',
  },
  {
    date: 'Sep 30, 2022; 20:32pm',
    jobs: '9 Jobs Total',
    runtime: '4 hours',
    id: 'b182de2b66994db6adc0925b5eb966f1',
  },
];

export const Default = () => {
  return (
    <Table>
      <Table.Head>
        <Table.Row>
          <Table.HeaderCell>Name</Table.HeaderCell>
          <Table.HeaderCell>Points</Table.HeaderCell>
        </Table.Row>
      </Table.Head>

      <Table.Body>
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

const iconItems = [
  {
    id: 'copy',
    content: 'Copy ID to DAG',
    closeOnClick: true,
  },
];

export const CheckboxRowTable = () => {
  const [rows, selectRows] = useState<string[]>([]);
  const addSelection = (value: string) => {
    rows.includes(value)
      ? selectRows(rows.filter((v) => v !== value))
      : selectRows([...rows, value]);
  };
  const allSelected = datumView.reduce(
    (acc, val) => acc && rows.includes(val.id),
    true,
  );
  const selectAll = () =>
    !allSelected ? selectRows(datumView.map((j) => j.id)) : selectRows([]);

  return (
    <div style={{overflow: 'scroll'}}>
      <p style={{height: 100}}>Selected Rows: {rows.toString()}</p>
      <Table>
        <Table.Head
          sticky
          hasCheckbox
          onClick={selectAll}
          isSelected={allSelected}
        >
          <Table.Row>
            <Table.HeaderCell>Started</Table.HeaderCell>
            <Table.HeaderCell>Run Progress</Table.HeaderCell>
            <Table.HeaderCell>Runtime</Table.HeaderCell>
            <Table.HeaderCell>ID</Table.HeaderCell>
            <Table.HeaderCell />
          </Table.Row>
        </Table.Head>

        <Table.Body>
          {datumView.map((job) => (
            <Table.Row
              key={job.id}
              onClick={() => addSelection(job.id)}
              isSelected={rows.includes(job.id)}
              hasCheckbox
              overflowMenuItems={iconItems}
            >
              <Table.DataCell>{job.date}</Table.DataCell>
              <Table.DataCell>{job.jobs}</Table.DataCell>
              <Table.DataCell>{job.runtime}</Table.DataCell>
              <Table.DataCell>{job.id}</Table.DataCell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    </div>
  );
};

export const SortableHeader = () => {
  const nameComparator = {
    name: 'Name',
    func: stringComparator,
    accessor: (record: Player) => record.name,
  };

  const pointsComparator = {
    name: 'Points',
    func: numberComparator,
    accessor: (record: Player) => record.points,
  };

  const {sortedData, setComparator, reversed, comparatorName} = useSort({
    data: bruinsPointLeaders,
    initialSort: nameComparator,
    initialDirection: 1,
  });

  return (
    <div style={{maxHeight: 330, overflow: 'scroll'}}>
      <Table>
        <Table.Head sticky>
          <Table.Row>
            <Table.HeaderCell
              onClick={() => setComparator(nameComparator)}
              sortable={true}
              sortLabel="name"
              sortSelected={comparatorName === 'Name'}
              sortReversed={!reversed}
            >
              Name
            </Table.HeaderCell>
            <Table.HeaderCell
              onClick={() => setComparator(pointsComparator)}
              sortable={true}
              sortLabel="points"
              sortSelected={comparatorName === 'Points'}
              sortReversed={!reversed}
            >
              Points
            </Table.HeaderCell>
          </Table.Row>
        </Table.Head>

        <Table.Body>
          {sortedData.map((player) => (
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
