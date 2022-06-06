import React, {useState} from 'react';

import {HamburgerSVG, EducationSVG, SupportSVG} from 'Svg';

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

export const Disabled = () => {
  return (
    <DefaultDropdown items={items} buttonOpts={{disabled: true}}>
      Difficulty
    </DefaultDropdown>
  );
};

export const SideOpen = () => {
  return (
    <DefaultDropdown items={items} storeSelected sideOpen>
      Difficulty
    </DefaultDropdown>
  );
};

const iconItems: DropdownItem[] = [
  {
    id: 'start',
    content: 'Start Tutorial',
    IconSVG: EducationSVG,
    buttonStyle: 'tertiary',
  },
  {
    id: 'support',
    content: 'Contact Support',
    IconSVG: SupportSVG,
    buttonStyle: 'tertiary',
  },
];

export const Icon = () => {
  return (
    <div style={{backgroundColor: 'black', padding: '1rem'}}>
      <DefaultDropdown
        items={iconItems}
        storeSelected
        buttonOpts={{
          hideChevron: true,
          IconSVG: HamburgerSVG,
          buttonType: 'tertiary',
        }}
        menuOpts={{pin: 'left'}}
      />
    </div>
  );
};
