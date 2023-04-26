import classnames from 'classnames';
import React from 'react';

import {UUID_WITHOUT_DASHES_REGEX} from '@dash-frontend/constants/pachCore';
import {
  Icon,
  Input,
  Form,
  Label,
  Button,
  CheckmarkSVG,
  FilterSVG,
  ChevronDownSVG,
} from '@pachyderm/components';

import BrandedDocLink from '../BrandedDocLink';

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
          buttonType={globalIdFilter ? 'primary' : 'ghost'}
          color={globalIdFilter && 'black'}
          onClick={() => setDropdownOpen(!dropdownOpen)}
          className={classnames(styles.filter, {
            [styles.active]: globalIdFilter,
          })}
          IconSVG={!globalIdFilter ? FilterSVG : undefined}
        >
          {globalIdFilter && <div className={styles.dot} />}
          {!globalIdFilter ? (
            'Filter by Global ID'
          ) : (
            <div className={styles.buttonContent}>
              Global ID: <b>{`${globalIdFilter.slice(0, 6)}...`}</b>
              <Icon small color="white">
                <ChevronDownSVG />
              </Icon>
            </div>
          )}
        </Button>
        {dropdownOpen && (
          <div className={styles.dropdownBase}>
            <div className={styles.infoText}>
              Filtering by Global ID will show you all the commits and jobs with
              the same ID.{' '}
              <BrandedDocLink
                pathWithoutDomain="concepts/advanced-concepts/globalid/"
                className={styles.link}
              >
                Read More
              </BrandedDocLink>
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
                    ? 'Clear Global ID filter'
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
