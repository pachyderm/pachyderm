import BodyContent from './components/BodyContent';
import BodyHeader from './components/BodyHeader';
import BodyHeaderDropdown from './components/BodyHeaderDropdown';
import BodyHeaderTabs from './components/BodyHeaderTabs';
import TableViewBody from './TableViewBody';

export default Object.assign(TableViewBody, {
  Header: BodyHeader,
  Dropdown: BodyHeaderDropdown,
  Tabs: BodyHeaderTabs,
  Content: BodyContent,
});
