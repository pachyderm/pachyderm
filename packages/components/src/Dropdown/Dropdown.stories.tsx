import React, {useState} from 'react';

import {SearchableDropdown} from './Dropdown';

import {DefaultDropdown, DropdownItem} from './';

export default {title: 'Dropdown'};

const items: DropdownItem[] = [
  {id: 'easy', content: 'Easy'},
  {id: 'medium', content: 'Medium', disabled: true},
  {id: 'hard', content: 'Hard'},
];

export const Default = () => {
  return (
    <DefaultDropdown items={items} storeSelected>
      Difficulty
    </DefaultDropdown>
  );
};

export const CloseOnClick = () => {
  const closeOnClickItems: DropdownItem[] = items.map((item) => ({
    ...item,
    closeOnClick: true,
  }));

  return (
    <DefaultDropdown items={closeOnClickItems} storeSelected>
      Difficulty
    </DefaultDropdown>
  );
};

export const Search = () => {
  const [branch, setBranch] = useState('master');

  const items: DropdownItem[] = [
    {id: 'master', value: 'master', content: 'master', closeOnClick: true},
    {id: 'staging', value: 'staging', content: 'staging', closeOnClick: true},
    {
      id: 'development',
      value: 'development',
      content: 'development',
      closeOnClick: true,
    },
  ];

  return (
    <SearchableDropdown
      selectedId={branch}
      onSelect={setBranch}
      items={items}
      searchOpts={{placeholder: 'Search a branch by name'}}
      emptyResultsContent={'No matching branches found.'}
    >
      Commit (Branch: {branch})
    </SearchableDropdown>
  );
};
