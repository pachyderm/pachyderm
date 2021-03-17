package testing

import (
	"fmt"
	"testing"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

const (
	baseCollectionSize = 10
	originalValue      = "old"
	changedValue       = "new"
	collectionName     = "test_items"
)

type WriteCallback func(func(col.ReadWriteCollection) error) error

// populateCollection writes a standard set of rows to the collection for use by
// individual tests.
func populateCollection(rw col.ReadWriteCollection) error {
	for i := 0; i < baseCollectionSize; i++ {
		testProto := makeProto(i)
		testProto.Value = originalValue
		if err := rw.Create(testProto.ID, testProto); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to turn an int ID into a string so we don't need to use string literals
func makeID(i int) string {
	return fmt.Sprintf("%d", i)
}

// Helper function to instantiate the proto for a row by ID
func makeProto(i int) *TestItem {
	return &TestItem{ID: makeID(i), Value: changedValue}
}

type RowDiff struct {
	Deleted []int
	Changed []int
	Added   []int
}

// checkCollection validates that the contents of the collection match the
// expected set of rows.
func checkCollection(t *testing.T, ro col.ReadOnlyCollection, diff RowDiff) {
	expectedRows := map[string]string{}
	for i := 0; i < baseCollectionSize; i++ {
		expectedRows[makeID(i)] = originalValue
	}
	if diff.Deleted != nil {
		for _, i := range diff.Deleted {
			_, ok := expectedRows[makeID(i)]
			require.True(t, ok, "test specified a deleted row that was not in the original set")
			delete(expectedRows, makeID(i))
		}
	}
	if diff.Changed != nil {
		for _, i := range diff.Changed {
			_, ok := expectedRows[makeID(i)]
			require.True(t, ok, "test specified a changed row that was not in the original set")
			expectedRows[makeID(i)] = changedValue
		}
	}
	if diff.Added != nil {
		for _, i := range diff.Added {
			_, ok := expectedRows[makeID(i)]
			require.False(t, ok, "test specified an added row that was already in the original set")
			expectedRows[makeID(i)] = changedValue
		}
	}

	testProto := &TestItem{}
	actualRows := map[string]string{}
	require.NoError(t, ro.List(testProto, col.DefaultOptions(), func() error {
		actualRows[testProto.ID] = testProto.Value
		return nil
	}))

	for k, v := range expectedRows {
		other, ok := actualRows[k]
		require.True(t, ok, "row '%s' was expected to be present but was not found", k)
		require.Equal(t, v, other, "row '%s' had an unexpected value", k)
	}

	for k := range actualRows {
		_, ok := expectedRows[k]
		require.True(t, ok, "row '%s' was present but was not expected", k)
	}
}

func readOnlyTests(
	suite *testing.T,
	newCollection func(*testing.T) (col.ReadOnlyCollection, WriteCallback),
) {
	suite.Run("ReadOnly", func(suite *testing.T) {
		emptyCol, _ := newCollection(suite)
		staticCol, writer := newCollection(suite)
		require.NoError(suite, writer(populateCollection))

		suite.Run("Get", func(subsuite *testing.T) {
			subsuite.Run("ErrNotFound", func(t *testing.T) {
				itemProto := &TestItem{}
				err := staticCol.Get("baz", itemProto)
				require.True(t, errors.Is(err, col.ErrNotFound{}))
				require.True(t, col.IsErrNotFound(err))
			})
			subsuite.Run("Success", func(t *testing.T) {
				itemProto := &TestItem{}
				require.NoError(t, staticCol.Get("5", itemProto))
				require.Equal(t, "5", itemProto.ID)
			})
		})

		suite.Run("GetByIndex", func(subsuite *testing.T) {
			subsuite.Run("ErrNotFound", func(t *testing.T) {
			})
			subsuite.Run("Success", func(t *testing.T) {
			})
		})

		suite.Run("List", func(subsuite *testing.T) {
			subsuite.Run("Empty", func(t *testing.T) {
			})
			subsuite.Run("Success", func(t *testing.T) {
			})
		})

		suite.Run("ListPrefix", func(subsuite *testing.T) {
			subsuite.Run("Empty", func(t *testing.T) {
			})
			subsuite.Run("Success", func(t *testing.T) {
			})
		})

		suite.Run("Count", func(subsuite *testing.T) {
			subsuite.Run("Success", func(t *testing.T) {
				count, err := staticCol.Count()
				require.NoError(t, err)
				require.Equal(t, int64(10), count)

				count, err = emptyCol.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})
		})

		suite.Run("Watch", func(subsuite *testing.T) {
			subsuite.Run("Success", func(t *testing.T) {
			})
		})

		suite.Run("WatchF", func(subsuite *testing.T) {
			subsuite.Run("Success", func(t *testing.T) {
			})
		})

		suite.Run("WatchOne", func(subsuite *testing.T) {
			subsuite.Run("ErrNotFound", func(t *testing.T) {
			})
			subsuite.Run("Success", func(t *testing.T) {
			})
		})

		suite.Run("WatchOneF", func(subsuite *testing.T) {
			subsuite.Run("ErrNotFound", func(t *testing.T) {
			})
			subsuite.Run("Success", func(t *testing.T) {
			})
		})

		suite.Run("WatchByIndex", func(subsuite *testing.T) {
			subsuite.Run("Success", func(t *testing.T) {
			})
		})
	})
}

func readWriteTests(
	suite *testing.T,
	newCollection func(t *testing.T) (col.ReadOnlyCollection, WriteCallback),
) {
	initCollection := func(t *testing.T) (col.ReadOnlyCollection, WriteCallback) {
		readOnly, writer := newCollection(t)
		require.NoError(t, writer(populateCollection))
		return readOnly, writer
	}

	suite.Run("ReadWrite", func(suite *testing.T) {
		suite.Run("Create", func(subsuite *testing.T) {
			subsuite.Run("Success", func(t *testing.T) {
				newID := 10
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := makeProto(newID)
					return rw.Create(testProto.ID, testProto)
				})
				require.NoError(t, err)
				checkCollection(t, readOnly, RowDiff{Added: []int{newID}})
			})

			subsuite.Run("ErrExists", func(t *testing.T) {
				// Fails when overwriting a row
				overwriteID := 5
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := makeProto(overwriteID)
					return rw.Create(testProto.ID, testProto)
				})
				require.YesError(t, err)
				require.True(t, col.IsErrExists(err))
				require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: makeID(overwriteID)}))
				checkCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(t *testing.T) {
				// Fails when doing a successful create and an overwrite in the same transaction
				overwriteID := 6
				newID := 10
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := makeProto(newID)
					if err := rw.Create(testProto.ID, testProto); err != nil {
						return err
					}
					testProto = makeProto(overwriteID)
					return rw.Create(testProto.ID, testProto)
				})
				require.YesError(t, err)
				require.True(t, col.IsErrExists(err))
				require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: makeID(overwriteID)}))
				checkCollection(t, readOnly, RowDiff{})
			})
		})

		suite.Run("Put", func(subsuite *testing.T) {
			subsuite.Run("Success", func(t *testing.T) {
			})
			subsuite.Run("ErrExists", func(t *testing.T) {
				// Fails if row already exists
				overwriteID := 5
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := makeProto(overwriteID)
					return rw.Create(testProto.ID, testProto)
				})
				require.YesError(t, err)
				require.True(t, col.IsErrExists(err))
				require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: makeID(overwriteID)}))
				checkCollection(t, readOnly, RowDiff{})
			})
		})
	})
}
