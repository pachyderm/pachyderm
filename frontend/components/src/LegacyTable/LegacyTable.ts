import Body from './components/Body';
import DataCell from './components/DataCell';
import Head from './components/Head';
import HeaderCell from './components/HeaderCell';
import Row from './components/Row';
import LegacyTable from './components/Table';

export default Object.assign(LegacyTable, {
  Head,
  Row,
  HeaderCell,
  DataCell,
  Body,
});
