import TableViewBody from './components/TableViewBody';
import TableViewHeader from './components/TableViewHeader';
import TableView from './TableView';

export {TableViewBody, TableViewHeader};

export default Object.assign(TableView, {
  Body: TableViewBody,
  Header: TableViewHeader,
});
