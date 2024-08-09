import {render, waitFor, screen, act, fireEvent} from '@testing-library/react';
import React from 'react';
import {useForm} from 'react-hook-form';

import {click, paste, type} from '@dash-frontend/testHelpers';
import {Form} from '@pachyderm/components';

import {Keys} from '../../lib/types';
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
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');

    await type(input, 'hello');
    expect(input).toHaveValue('hello');
    await type(input, ',');
    expect(input).toHaveValue('');

    const helloTag = screen.getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should show a tag after space press', async () => {
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');

    await type(input, 'hello');
    expect(input).toHaveValue('hello');
    await type(input, ' ');
    expect(input).toHaveValue('');

    const helloTag = screen.getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should show a tag after tab press', async () => {
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');

    await type(input, 'hello');
    expect(input).toHaveValue('hello');

    // userEvent tab doesn't work here anymore
    act(() => fireEvent.keyDown(input, {key: Keys.Tab}));
    expect(input).toHaveValue('');

    const helloTag = screen.getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should show a tag after Enter press', async () => {
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');

    await type(input, 'hello');
    expect(input).toHaveValue('hello');
    await type(input, '{enter}');
    expect(input).toHaveValue('');

    const helloTag = screen.getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should remove tag via close button', async () => {
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');

    await type(input, 'hello');
    await type(input, ' ');
    await type(input, 'world');
    await type(input, ' ');

    expect(screen.getByText('hello')).toBeInTheDocument();
    expect(screen.getByText('world')).toBeInTheDocument();

    const removeHelloButton = screen.getByLabelText('Remove hello');

    await click(removeHelloButton);
    expect(screen.queryByText('hello')).not.toBeInTheDocument();
    expect(screen.getByText('world')).toBeInTheDocument();
  });

  it('should remove tag via backspace', async () => {
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');

    await type(input, 'hello');
    await type(input, ' ');
    await type(input, 'world');
    await type(input, ' ');

    expect(screen.getByText('hello')).toBeInTheDocument();
    expect(screen.getByText('world')).toBeInTheDocument();

    await type(input, '{backspace}');
    expect(screen.getByText('hello')).toBeInTheDocument();
    expect(screen.queryByText('world')).not.toBeInTheDocument();
  });

  it('should render tags for default values', () => {
    const defaultValues = ['hello'];
    render(<TagsInput defaultValues={defaultValues} />);
    const helloTag = screen.getByText('hello');

    expect(helloTag).toBeInTheDocument();
    expect(helloTag).toHaveClass('tag');
  });

  it('should set and clear focus', () => {
    render(<TagsInput />);
    const container = screen.getByTestId('TagsInput__container');
    const input = screen.getByTestId('TagsInput__input');

    expect(container).not.toHaveClass('focused');
    act(() => input.focus());
    expect(container).toHaveClass('focused');
    act(() => input.blur());
    expect(container).not.toHaveClass('focused');
  });

  it('should validate tags', async () => {
    const validate = (v: string) => v === 'world';
    render(<TagsInput validate={validate} />);
    const container = screen.getByTestId('TagsInput__container');
    const input = screen.getByTestId('TagsInput__input');

    await type(input, 'hello');
    await type(input, ' ');
    await type(input, 'world');
    await type(input, ' ');

    await waitFor(() => {
      expect(container).toHaveClass('errors');
    });
    expect(screen.queryByText('hello')).toHaveClass('tag');
    expect(screen.queryByText('hello')).toHaveClass('tagError');
    expect(screen.queryByText('world')).toHaveClass('tag');
    expect(screen.queryByText('world')).not.toHaveClass('tagError');

    const removeHelloButton = screen.getByLabelText('Remove hello');

    await click(removeHelloButton);
    await waitFor(() => {
      expect(container).not.toHaveClass('errors');
    });
  });

  it('should show and hide aria-describedby ids', async () => {
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');

    expect(input).toHaveAttribute('aria-describedby', '');

    await type(input, 'hello');
    await type(input, ' ');
    expect(screen.getByText('hello')).toHaveAttribute('id', 'tag-hello-0');
    expect(input).toHaveAttribute('aria-describedby', 'tag-hello-0');

    await type(input, 'world');
    await type(input, ' ');
    expect(screen.getByText('world')).toHaveAttribute('id', 'tag-world-1');
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
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');

    act(() => input.focus());

    await paste(
      'test1@test.com,test2@test.com test3@test.com, test4@test.com,,test5@test.com',
    );
    await type(input, '{enter}');

    const testTag1 = screen.getByText('test1@test.com');
    expect(testTag1).toBeInTheDocument();
    expect(testTag1).toHaveClass('tag');

    const testTag2 = screen.getByText('test2@test.com');
    expect(testTag2).toBeInTheDocument();
    expect(testTag2).toHaveClass('tag');

    const testTag3 = screen.getByText('test3@test.com');
    expect(testTag3).toBeInTheDocument();
    expect(testTag3).toHaveClass('tag');

    const testTag4 = screen.getByText('test4@test.com');
    expect(testTag4).toBeInTheDocument();
    expect(testTag4).toHaveClass('tag');

    const testTag5 = screen.getByText('test5@test.com');
    expect(testTag5).toBeInTheDocument();
    expect(testTag5).toHaveClass('tag');
  });

  it('should show a clear button after 2 entries', async () => {
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');
    act(() => input.focus());

    await paste('hello');
    await type(input, '{enter}');

    let clearButton = screen.queryByText('Clear Tagsinput');
    expect(clearButton).not.toBeInTheDocument();

    act(() => input.focus());
    await paste('world');
    await type(input, '{enter}');

    clearButton = screen.queryByText('Clear Tagsinput');
    expect(clearButton).toBeInTheDocument();
  });

  it('should clear form input on clear button press', async () => {
    render(<TagsInput />);
    const input = screen.getByTestId('TagsInput__input');
    act(() => input.focus());

    await paste('hello,world');
    await type(input, '{enter}');

    const helloTag = screen.getByText('hello');
    const worldTag = screen.getByText('world');

    expect(helloTag).toBeInTheDocument();
    expect(worldTag).toBeInTheDocument();

    const clearButton = screen.getByText('Clear Tagsinput');
    await click(clearButton);

    const text = screen.queryByText('hello');
    expect(text).not.toBeInTheDocument();
  });
});
