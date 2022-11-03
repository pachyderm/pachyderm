import {render, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {useForm} from 'react-hook-form';

import {Form} from 'Form';
import {click, paste, type} from 'testHelpers';

import {TagsInput as TagsInputComponent} from '../TagsInput';

const TagsInput = ({...props}) => {
  const formContext = useForm();

  return (
    <Form formContext={formContext}>
      <TagsInputComponent {...props} />
    </Form>
  );
};

describe('components/TagsInput', () => {
  it('should show a tag after comma press', async () => {
    const {getByTestId, getByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    await type(input, 'hello');
    expect(input).toHaveValue('hello');
    await type(input, ',');
    expect(input).toHaveValue('');

    const helloTag = getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should show a tag after space press', async () => {
    const {getByTestId, getByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    await type(input, 'hello');
    expect(input).toHaveValue('hello');
    await type(input, ' ');
    expect(input).toHaveValue('');

    const helloTag = getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should show a tag after tab press', async () => {
    const {getByTestId, getByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    await type(input, 'hello');
    expect(input).toHaveValue('hello');
    userEvent.tab();
    expect(input).toHaveValue('');

    const helloTag = getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should show a tag after Enter press', async () => {
    const {getByTestId, getByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    await type(input, 'hello');
    expect(input).toHaveValue('hello');
    await type(input, '{enter}');
    expect(input).toHaveValue('');

    const helloTag = getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should remove tag via close button', async () => {
    const {getByTestId, getByLabelText, queryByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    await type(input, 'hello');
    await type(input, ' ');
    await type(input, 'world');
    await type(input, ' ');

    expect(queryByText('hello')).toBeInTheDocument();
    expect(queryByText('world')).toBeInTheDocument();

    const removeHelloButton = getByLabelText('Remove hello');

    click(removeHelloButton);
    expect(queryByText('hello')).not.toBeInTheDocument();
    expect(queryByText('world')).toBeInTheDocument();
  });

  it('should remove tag via backspace', async () => {
    const {getByTestId, queryByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    await type(input, 'hello');
    await type(input, ' ');
    await type(input, 'world');
    await type(input, ' ');

    expect(queryByText('hello')).toBeInTheDocument();
    expect(queryByText('world')).toBeInTheDocument();

    await type(input, '{backspace}');
    expect(queryByText('hello')).toBeInTheDocument();
    expect(queryByText('world')).not.toBeInTheDocument();
  });

  it('should render tags for default values', () => {
    const defaultValues = ['hello'];
    const {getByText} = render(<TagsInput defaultValues={defaultValues} />);
    const helloTag = getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should set and clear focus', () => {
    const {getByTestId} = render(<TagsInput />);
    const container = getByTestId('TagsInput__container');
    const input = getByTestId('TagsInput__input');

    expect(container).not.toHaveClass('focused');
    input.focus();
    expect(container).toHaveClass('focused');
    input.blur();
    expect(container).not.toHaveClass('focused');
  });

  it('should validate tags', async () => {
    const validate = (v: string) => v === 'world';
    const {getByLabelText, getByTestId, queryByText} = render(
      <TagsInput validate={validate} />,
    );
    const container = getByTestId('TagsInput__container');
    const input = getByTestId('TagsInput__input');

    await type(input, 'hello');
    await type(input, ' ');
    await type(input, 'world');
    await type(input, ' ');

    await waitFor(() => {
      expect(container).toHaveClass('errors');
    });
    expect(queryByText('hello')).toHaveClass('tag');
    expect(queryByText('hello')).toHaveClass('tagError');
    expect(queryByText('world')).toHaveClass('tag');
    expect(queryByText('world')).not.toHaveClass('tagError');

    const removeHelloButton = getByLabelText('Remove hello');

    click(removeHelloButton);
    await waitFor(() => {
      expect(container).not.toHaveClass('errors');
    });
  });

  it('should show and hide aria-describedby ids', async () => {
    const {getByTestId, getByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    expect(input).toHaveAttribute('aria-describedby', '');

    await type(input, 'hello');
    await type(input, ' ');
    expect(getByText('hello')).toHaveAttribute('id', 'tag-hello-0');
    expect(input).toHaveAttribute('aria-describedby', 'tag-hello-0');

    await type(input, 'world');
    await type(input, ' ');
    expect(getByText('world')).toHaveAttribute('id', 'tag-world-1');
    expect(input).toHaveAttribute(
      'aria-describedby',
      'tag-hello-0 tag-world-1',
    );

    await type(input, '{backspace}');
    expect(input).toHaveAttribute('aria-describedby', 'tag-hello-0');
    await type(input, '{backspace}');
    expect(input).toHaveAttribute('aria-describedby', '');
  });

  it('should show tags for a pasted list of tags separated by commas and spaces', async () => {
    const {getByTestId, getByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    paste(
      input,
      'test1@test.com,test2@test.com test3@test.com, test4@test.com,,test5@test.com',
    );
    await type(input, '{enter}');

    const testTag1 = getByText('test1@test.com');
    expect(testTag1).toBeInTheDocument();
    expect(testTag1).toHaveClass('tag');

    const testTag2 = getByText('test2@test.com');
    expect(testTag2).toBeInTheDocument();
    expect(testTag2).toHaveClass('tag');

    const testTag3 = getByText('test3@test.com');
    expect(testTag3).toBeInTheDocument();
    expect(testTag3).toHaveClass('tag');

    const testTag4 = getByText('test4@test.com');
    expect(testTag4).toBeInTheDocument();
    expect(testTag4).toHaveClass('tag');

    const testTag5 = getByText('test5@test.com');
    expect(testTag5).toBeInTheDocument();
    expect(testTag5).toHaveClass('tag');
  });

  it('should show a clear button after 2 entries', async () => {
    const {getByTestId, queryByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    paste(input, 'hello');
    await type(input, '{enter}');

    let clearButton = queryByText('Clear Tagsinput');
    expect(clearButton).not.toBeInTheDocument();

    paste(input, 'world');
    await type(input, '{enter}');

    clearButton = queryByText('Clear Tagsinput');
    expect(clearButton).toBeInTheDocument();
  });

  it('should clear form input on clear button press', async () => {
    const {getByTestId, getByText, queryByText} = render(<TagsInput />);
    const input = getByTestId('TagsInput__input');

    paste(input, 'hello,world');
    await type(input, '{enter}');

    const helloTag = getByText('hello');
    const worldTag = getByText('world');

    expect(helloTag).toBeInTheDocument();
    expect(worldTag).toBeInTheDocument();

    const clearButton = getByText('Clear Tagsinput');
    click(clearButton);

    const text = queryByText('hello');
    expect(text).not.toBeInTheDocument();
  });
});
