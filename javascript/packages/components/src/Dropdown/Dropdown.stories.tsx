import React from 'react';

import Dropdown from './';

export default {title: 'Dropdown'};

export const Default = () => {
  return (
    <Dropdown storeSelected>
      <Dropdown.Button>Difficulty</Dropdown.Button>

      <Dropdown.Menu>
        <Dropdown.MenuItem id="easy">Easy</Dropdown.MenuItem>
        <Dropdown.MenuItem id="medium">Medium</Dropdown.MenuItem>
        <Dropdown.MenuItem id="hard">Hard</Dropdown.MenuItem>
      </Dropdown.Menu>
    </Dropdown>
  );
};

export const CloseOnClick = () => {
  return (
    <Dropdown storeSelected>
      <Dropdown.Button>Difficulty</Dropdown.Button>

      <Dropdown.Menu>
        <Dropdown.MenuItem id="easy" closeOnClick>
          Easy
        </Dropdown.MenuItem>
        <Dropdown.MenuItem id="medium" closeOnClick>
          Medium
        </Dropdown.MenuItem>
        <Dropdown.MenuItem id="hard" closeOnClick>
          Hard
        </Dropdown.MenuItem>
      </Dropdown.Menu>
    </Dropdown>
  );
};
