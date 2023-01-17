import {render, fireEvent, waitFor, screen} from '@testing-library/react';
import React from 'react';
import {useForm} from 'react-hook-form';

import {click} from '@dash-frontend/testHelpers';

import {Button} from '../../Button';
import {Form} from '../../Form';
import {Label} from '../../Label';
import {Select} from '../../Select';

describe('Select', () => {
  const submitSpy = jest.fn();

  const TestBed = () => {
    const formContext = useForm();

    return (
      <Form formContext={formContext} onSubmit={(values) => submitSpy(values)}>
        <Label htmlFor="select" id="select" label="Label" />
        <Select id="select">
          <Select.Option value="1">Option 1</Select.Option>
          <Select.Option value="2">Option 2</Select.Option>
        </Select>

        <Button type="submit">Button</Button>
      </Form>
    );
  };

  it('should allow users to select an option via click', async () => {
    render(<TestBed />);

    const comboBox = screen.getByRole('combobox');
    const submitButton = screen.getByRole('button');
    const options = screen.getAllByRole('option');

    await click(comboBox);

    await click(options[0]);
    await click(submitButton);

    await waitFor(() =>
      expect(submitSpy).toHaveBeenCalledWith(
        expect.objectContaining({select: '1'}),
      ),
    );
  });

  it('should allow users to select an option via keyboard', async () => {
    render(<TestBed />);

    const comboBox = screen.getByRole('combobox');
    const submitButton = screen.getByRole('button');

    fireEvent.keyDown(comboBox, {key: 'Enter'});

    fireEvent.keyDown(comboBox, {key: 'ArrowDown'});
    fireEvent.keyDown(comboBox, {key: 'ArrowDown'});
    fireEvent.keyDown(comboBox, {key: ' '});

    await click(submitButton);

    await waitFor(() =>
      expect(submitSpy).toHaveBeenCalledWith(
        expect.objectContaining({select: '2'}),
      ),
    );
  });
});
