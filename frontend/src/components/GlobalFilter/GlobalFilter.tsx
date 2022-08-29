import {
  ExternalLinkSVG,
  Icon,
  Link,
  Input,
  Form,
  Label,
  Button,
  CheckmarkSVG,
} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';

import {UUID_WITHOUT_DASHES_REGEX} from '@dash-frontend/constants/pachCore';

import styles from './GlobalFilter.module.css';
import useGlobalFilter from './hooks/useGlobalFilter';

const GlobalFilter: React.FC = () => {
  const {
    containerRef,
    formCtx,
    dropdownOpen,
    setDropdownOpen,
    loading,
    globalIdFilter,
    globalIdInput,
    handleSubmit,
  } = useGlobalFilter();

  return (
    <Form formContext={formCtx}>
      <div className={styles.base} ref={containerRef}>
        <Button
          buttonType="tertiary"
          onClick={() => setDropdownOpen(true)}
          className={classnames(styles.filter, {
            [styles.active]: globalIdFilter,
          })}
        >
          {globalIdFilter && <div className={styles.dot} />}
          {!globalIdFilter
            ? 'Filter by Global ID'
            : `Global ID: ${globalIdFilter.slice(0, 8)}`}
        </Button>
        {dropdownOpen && (
          <div className={styles.dropdownBase}>
            <div className={styles.infoText}>
              Filtering by Global ID will show you all the commits and jobs with
              the same ID.{' '}
              <Link
                externalLink
                to="https://docs.pachyderm.com/latest/concepts/advanced-concepts/globalid/"
                className={styles.link}
              >
                Read More
                <Icon small>
                  <ExternalLinkSVG aria-hidden />
                </Icon>
              </Link>
            </div>
            <div className={styles.form}>
              <Label htmlFor="globalId" label="Global ID" />
              {globalIdFilter === globalIdInput && (
                <Icon color="green" className={styles.checkMarkSVG}>
                  <CheckmarkSVG />
                </Icon>
              )}
              <Input
                placeholder="Copy and paste your Global ID"
                data-testid="GlobalFilter__name"
                type="text"
                id="globalId"
                name="globalId"
                spellCheck={false}
                validationOptions={{
                  pattern: {
                    value: UUID_WITHOUT_DASHES_REGEX,
                    message: 'Not a valid Global ID',
                  },
                }}
                clearable
                disabled={loading}
                autoFocus={true}
              />
              <Button
                disabled={loading}
                className={styles.submitButton}
                aria-label={
                  globalIdInput === globalIdFilter
                    ? 'Clear global ID filter'
                    : 'Apply Global ID Filter'
                }
                onClick={handleSubmit}
              >
                {globalIdInput === globalIdFilter ? 'Clear ID' : 'Apply'}
              </Button>
            </div>
          </div>
        )}
      </div>
    </Form>
  );
};

export default GlobalFilter;
