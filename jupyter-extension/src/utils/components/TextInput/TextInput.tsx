import React from 'react';

export type TextInputProps = {
  onSubmit: (value: string) => void;
  placeholder?: string;
  testIdPrefix?: string;
};

export const TextInput: React.FC<TextInputProps> = ({
  onSubmit,
  placeholder,
  testIdPrefix,
}) => {
  const [commit, setCommit] = React.useState('');

  return (
    <div className="pachyderm-TextInput">
      <div className="bp3-input-group jp-InputGroup">
        <input
          type="text"
          className="bp3-input"
          placeholder={placeholder}
          value={commit}
          onSubmit={() => {
            onSubmit(commit);
          }}
          onKeyPress={(e) => {
            if (e.key === 'Enter') {
              onSubmit(commit);
            }
          }}
          onChange={(e) => {
            setCommit(e.target.value);
          }}
          data-testid={`${testIdPrefix}TextInput-input`}
        />
      </div>
    </div>
  );
};

export default TextInput;
