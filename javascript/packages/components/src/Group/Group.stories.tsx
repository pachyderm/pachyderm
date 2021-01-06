import React from 'react';

import Group, {GroupProps} from './Group';
export default {
  title: 'Group',
};

const Item = ({text}: {text: string}) => (
  <div style={{backgroundColor: 'grey', width: '5rem', height: '5rem'}}>
    {text}
  </div>
);

const Items = () => (
  <>
    <Item text="Item_1" />
    <Item text="Item_2" />
    <Item text="Item_3" />
    <Item text="Item_4" />
  </>
);

export const Default = (props: GroupProps) => {
  return (
    <Group {...props}>
      <Items />
    </Group>
  );
};

export const Horizontal = () => {
  return <Default spacing={32} justify="center" />;
};

export const Vertical = () => {
  return <Default spacing={32} align="center" vertical />;
};

export const Divider = () => {
  return (
    <>
      <h1> Horizontal Divider</h1>
      <Group spacing={32}>
        <Item text="Item_1" />
        <Group.Divider />
        <Item text="Item_2" />
      </Group>
      <br />
      <h1> Vertical Divider</h1>
      <Group spacing={32} vertical>
        <Item text="Item_1" />
        <Group.Divider />
        <Item text="Item_2" />
      </Group>
    </>
  );
};
