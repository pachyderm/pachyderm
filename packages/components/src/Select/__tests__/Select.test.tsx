import {render, fireEvent, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {useForm} from 'react-hook-form';

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

  const renderTestBed = () => {
    return render(<TestBed />);
  };

  it('should allow users to select an option via click', async () => {
    const {getAllByRole, getByRole} = renderTestBed();

    const comboBox = getByRole('combobox');
    const submitButton = getByRole('button');
    const options = getAllByRole('option');

    userEvent.click(comboBox);

    userEvent.click(options[0]);
    userEvent.click(submitButton);

    await waitFor(() =>
      expect(submitSpy).toHaveBeenCalledWith(
        expect.objectContaining({select: '1'}),
      ),
    );
  });

  it('should allow users to select an option via keyboard', async () => {
    const {getByRole} = renderTestBed();

    const comboBox = getByRole('combobox');
    const submitButton = getByRole('button');

    fireEvent.keyDown(comboBox, {key: 'Enter'});

    fireEvent.keyDown(comboBox, {key: 'ArrowDown'});
    fireEvent.keyDown(comboBox, {key: 'ArrowDown'});
    fireEvent.keyDown(comboBox, {key: ' '});

    userEvent.click(submitButton);

    await waitFor(() =>
      expect(submitSpy).toHaveBeenCalledWith(
        expect.objectContaining({select: '2'}),
      ),
    );
  });
});
