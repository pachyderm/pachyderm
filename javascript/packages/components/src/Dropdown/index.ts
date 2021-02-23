import DropdownButton from './components/DropdownButton';
import DropdownMenu from './components/DropdownMenu';
import DropdownMenuItem from './components/DropdownMenuItem';
import Dropdown from './Dropdown';
import useDropdown from './hooks/useDropdown';

export {useDropdown};

export default Object.assign(Dropdown, {
  Button: DropdownButton,
  Menu: DropdownMenu,
  MenuItem: DropdownMenuItem,
});
