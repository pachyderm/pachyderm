package testing

import (
	"fmt"
	"testing"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

const (
	defaultCollectionSize = 10
	originalValue         = "old"
	changedValue          = "new"
	collectionName        = "test_items"
)

var (
	TestPrimaryIndex   = &col.Index{Field: "ID"}
	TestSecondaryIndex = &col.Index{Field: "Value"}
)

type TestError struct{}

func (te TestError) Is(other error) bool {
	_, ok := other.(TestError)
	return ok
}

func (te TestError) Error() string {
	return "TestError"
}

type WriteCallback func(func(col.ReadWriteCollection) error) error

func intRange(start int, end int) []int {
	result := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, i)
	}
	return result
}

// populateCollection writes a standard set of rows to the collection for use by
// individual tests.
func populateCollection(rw col.ReadWriteCollection) error {
	for i := 0; i < defaultCollectionSize; i++ {
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

// checkDefaultCollection validates that the contents of the collection match
// the expected set of rows based on a diff from the default collection
// populated by populateCollection.
func checkDefaultCollection(t *testing.T, ro col.ReadOnlyCollection, diff RowDiff) {
	expected := map[string]string{}
	for i := 0; i < defaultCollectionSize; i++ {
		expected[makeID(i)] = originalValue
	}
	if diff.Deleted != nil {
		for _, i := range diff.Deleted {
			_, ok := expected[makeID(i)]
			require.True(t, ok, "test specified a deleted row that was not in the original set")
			delete(expected, makeID(i))
		}
	}
	if diff.Changed != nil {
		for _, i := range diff.Changed {
			_, ok := expected[makeID(i)]
			require.True(t, ok, "test specified a changed row that was not in the original set")
			expected[makeID(i)] = changedValue
		}
	}
	if diff.Added != nil {
		for _, i := range diff.Added {
			_, ok := expected[makeID(i)]
			require.False(t, ok, "test specified an added row that was already in the original set")
			expected[makeID(i)] = changedValue
		}
	}
	checkCollection(t, ro, expected)
}

func checkItem(t *testing.T, item *TestItem, id string, value string) error {
	if item.ID != id {
		return errors.Errorf("Item has incorrect ID, expected: %s, actual: %s", id, item.ID)
	}
	if item.Value != value {
		return errors.Errorf("Item has incorrect Value, expected: %s, actual: %s", value, item.Value)
	}
	return nil
}

func checkCollection(t *testing.T, ro col.ReadOnlyCollection, expected map[string]string) {
	testProto := &TestItem{}
	actual := map[string]string{}
	require.NoError(t, ro.List(testProto, col.DefaultOptions(), func() error {
		actual[testProto.ID] = testProto.Value
		return nil
	}))

	for k, v := range expected {
		other, ok := actual[k]
		require.True(t, ok, "row '%s' was expected to be present but was not found", k)
		require.Equal(t, v, other, "row '%s' had an unexpected value", k)
	}

	for k := range actual {
		_, ok := expected[k]
		require.True(t, ok, "row '%s' was present but was not expected", k)
	}
}

func collectionTests(
	parent *testing.T,
	newCollection func(*testing.T) (col.ReadOnlyCollection, WriteCallback),
) {
	parent.Run("ReadOnly", func(suite *testing.T) {
		suite.Parallel()
		emptyCol, _ := newCollection(suite)
		defaultCol, writer := newCollection(suite)
		require.NoError(suite, writer(populateCollection))

		suite.Run("Get", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("ErrNotFound", func(t *testing.T) {
				t.Parallel()
				itemProto := &TestItem{}
				err := defaultCol.Get("baz", itemProto)
				require.True(t, errors.Is(err, col.ErrNotFound{}), "Incorrect error: %v", err)
				require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
			})
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				itemProto := &TestItem{}
				require.NoError(t, defaultCol.Get("5", itemProto))
				require.Equal(t, "5", itemProto.ID)
			})
		})

		suite.Run("GetByIndex", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("ErrNotFound", func(t *testing.T) {
				t.Parallel()
			})
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
			})
		})

		suite.Run("List", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Empty", func(t *testing.T) {
				t.Parallel()
				testProto := &TestItem{}
				err := emptyCol.List(testProto, col.DefaultOptions(), func() error {
					return errors.Errorf("List callback should not have been called for an empty collection")
				})
				require.NoError(t, err)
			})

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				items := map[string]string{}
				testProto := &TestItem{}
				err := defaultCol.List(testProto, col.DefaultOptions(), func() error {
					items[testProto.ID] = testProto.Value
					// Clear testProto.ID and testProto.Value just to make sure they get overwritten each time
					testProto.ID = ""
					testProto.Value = ""
					return nil
				})
				require.NoError(t, err)

				keys := []string{}
				for k, v := range items {
					keys = append(keys, k)
					require.Equal(t, originalValue, v)
				}
				expectedKeys := []string{}
				for i := 0; i < defaultCollectionSize; i++ {
					expectedKeys = append(expectedKeys, makeID(i))
				}
				require.ElementsEqual(t, expectedKeys, keys)
			})

			subsuite.Run("Sort", func(subsuite *testing.T) {
				subsuite.Parallel()

				testSort := func(t *testing.T, ro col.ReadOnlyCollection, target col.SortTarget, expectedAsc []string) {
					t.Parallel()

					collect := func(order col.SortOrder) []string {
						keys := []string{}
						testProto := &TestItem{}
						err := ro.List(testProto, &col.Options{Target: target, Order: order}, func() error {
							keys = append(keys, testProto.ID)
							return nil
						})
						require.NoError(t, err)
						return keys
					}

					t.Run("SortAscend", func(t *testing.T) {
						t.Parallel()
						keys := collect(col.SortAscend)
						require.Equal(t, expectedAsc, keys)
					})

					t.Run("SortDescend", func(t *testing.T) {
						t.Parallel()
						keys := collect(col.SortDescend)
						reversed := []string{}
						for i := len(expectedAsc) - 1; i >= 0; i-- {
							reversed = append(reversed, expectedAsc[i])
						}
						require.Equal(t, reversed, keys)
					})

					t.Run("SortNone", func(t *testing.T) {
						t.Parallel()
						keys := collect(col.SortNone)
						require.ElementsEqual(t, expectedAsc, keys)
					})
				}

				subsuite.Run("CreatedAt", func(t *testing.T) {
					// The defaultCol was written in a single transaction, so all the rows
					// have the same created time - make our own
					createCol, writer := newCollection(suite)

					createKeys := []string{"0", "6", "7", "9", "3", "8", "4", "1", "2", "5"}
					for _, k := range createKeys {
						err := writer(func(rw col.ReadWriteCollection) error {
							testProto := &TestItem{ID: k, Value: originalValue}
							return rw.Create(k, testProto)
						})
						require.NoError(t, err)
					}

					checkDefaultCollection(t, createCol, RowDiff{})
					testSort(t, createCol, col.SortByCreateRevision, createKeys)
				})

				subsuite.Run("UpdatedAt", func(t *testing.T) {
					// Create a new collection that we can modify to get a different ordering here
					modCol, writer := newCollection(suite)
					require.NoError(suite, writer(populateCollection))

					modKeys := []string{"5", "7", "2", "9", "1", "0", "8", "4", "3", "6"}
					for _, k := range modKeys {
						err := writer(func(rw col.ReadWriteCollection) error {
							testProto := &TestItem{}
							if err := rw.Update(k, testProto, func() error {
								testProto.Value = changedValue
								return nil
							}); err != nil {
								return err
							}
							return nil
						})
						require.NoError(t, err)
					}
					testSort(t, modCol, col.SortByModRevision, modKeys)
				})

				subsuite.Run("Key", func(t *testing.T) {
					testSort(t, defaultCol, col.SortByKey, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"})
				})
			})

			subsuite.Run("ErrorInCallback", func(t *testing.T) {
				t.Parallel()
				count := 0
				testProto := &TestItem{}
				err := defaultCol.List(testProto, col.DefaultOptions(), func() error {
					count++
					return &TestError{}
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				require.Equal(t, 1, count, "List callback was called multiple times despite erroring")
			})

			subsuite.Run("ErrBreak", func(t *testing.T) {
				t.Parallel()
				count := 0
				testProto := &TestItem{}
				err := defaultCol.List(testProto, col.DefaultOptions(), func() error {
					count++
					return errutil.ErrBreak
				})
				require.NoError(t, err)
				require.Equal(t, 1, count, "List callback was called multiple times despite aborting")
			})
		})

		suite.Run("Count", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				count, err := defaultCol.Count()
				require.NoError(t, err)
				require.Equal(t, int64(10), count)

				count, err = emptyCol.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})
		})

		suite.Run("Watch", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
			})
		})

		suite.Run("WatchF", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
			})
		})

		suite.Run("WatchOne", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("ErrNotFound", func(t *testing.T) {
				t.Parallel()
			})
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
			})
		})

		suite.Run("WatchOneF", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("ErrNotFound", func(t *testing.T) {
				t.Parallel()
			})
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
			})
		})

		suite.Run("WatchByIndex", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
			})
		})
	})

	parent.Run("ReadWrite", func(suite *testing.T) {
		suite.Parallel()
		initCollection := func(t *testing.T) (col.ReadOnlyCollection, WriteCallback) {
			readOnly, writer := newCollection(t)
			require.NoError(t, writer(populateCollection))
			return readOnly, writer
		}

		testRollback := func(t *testing.T, cb func(rw col.ReadWriteCollection) error) error {
			readOnly, writer := initCollection(t)
			err := writer(func(rw col.ReadWriteCollection) error {
				return cb(rw)
			})
			require.YesError(t, err)
			checkDefaultCollection(t, readOnly, RowDiff{})
			return err
		}

		suite.Run("Get", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				readOnly, writer := initCollection(t)

				err := writer(func(rw col.ReadWriteCollection) error {
					itemProto := &TestItem{}
					if err := rw.Get(makeID(8), itemProto); err != nil {
						return err
					}
					if itemProto.Value != originalValue {
						return errors.Errorf("Incorrect value when fetching via ReadWriteCollection.Get, expected: %s, actual: %s", originalValue, itemProto.Value)
					}
					return nil
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("AfterDelete", func(t *testing.T) {
				t.Parallel()
				readOnly, writer := initCollection(t)

				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := &TestItem{}
					if err := rw.Get(makeID(1), testProto); err != nil {
						return err
					}

					oobID := 10
					if err := rw.Get(makeID(oobID), testProto); !col.IsErrNotFound(err) {
						return errors.Wrapf(err, "Expected ErrNotFound for key '%d', but got", oobID)
					}

					rw.DeleteAll()

					for id := 0; id < defaultCollectionSize; id++ {
						if err := rw.Get(makeID(id), testProto); !col.IsErrNotFound(err) {
							return errors.Wrapf(err, "Expected ErrNotFound for key '%d', but got", id)
						}
					}
					return nil
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Deleted: intRange(0, 10)})

				count, err := readOnly.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})
		})

		suite.Run("Create", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				newID := 10
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := makeProto(newID)
					return rw.Create(testProto.ID, testProto)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Added: []int{newID}})
			})

			subsuite.Run("ErrExists", func(t *testing.T) {
				t.Parallel()
				overwriteID := 5
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := makeProto(overwriteID)
					return rw.Create(testProto.ID, testProto)
				})
				require.YesError(t, err)
				require.True(t, col.IsErrExists(err), "Incorrect error: %v", err)
				require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: makeID(overwriteID)}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()
				subsuite.Run("CreateError", func(t *testing.T) {
					t.Parallel()
					overwriteID := 6
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(10)
						if err := rw.Create(testProto.ID, testProto); err != nil {
							return err
						}
						testProto = makeProto(overwriteID)
						return rw.Create(testProto.ID, testProto)
					})
					require.True(t, col.IsErrExists(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: makeID(overwriteID)}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(10)
						if err := rw.Create(testProto.ID, testProto); err != nil {
							return err
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("Put", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Insert", func(t *testing.T) {
				t.Parallel()
				newID := 10
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := makeProto(newID)
					return rw.Put(testProto.ID, testProto)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Added: []int{newID}})
			})

			subsuite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				overwriteID := 5
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := makeProto(overwriteID)
					return rw.Put(testProto.ID, testProto)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Changed: []int{overwriteID}})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()
				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(10)
						if err := rw.Put(testProto.ID, testProto); err != nil {
							return err
						}
						testProto = makeProto(8)
						if err := rw.Put(testProto.ID, testProto); err != nil {
							return err
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("Update", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				updateID := 1
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := &TestItem{}
					if err := rw.Upsert(makeID(updateID), testProto, func() error {
						if err := checkItem(t, testProto, makeID(updateID), originalValue); err != nil {
							return err
						}
						testProto.Value = changedValue
						return nil
					}); err != nil {
						return err
					}
					if err := rw.Get(makeID(updateID), testProto); err != nil {
						return err
					}
					return checkItem(t, testProto, makeID(updateID), changedValue)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Changed: []int{updateID}})
			})

			subsuite.Run("ErrorInCallback", func(t *testing.T) {
				t.Parallel()
				updateID := 2
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := &TestItem{}
					return rw.Update(makeID(updateID), testProto, func() error {
						return &TestError{}
					})
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("NotFound", func(t *testing.T) {
				t.Parallel()
				notExistsID := 10
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := &TestItem{}
					return rw.Update(makeID(notExistsID), testProto, func() error {
						return nil
					})
				})
				require.YesError(t, err)
				require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
				require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: makeID(notExistsID)}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()
				subsuite.Run("UpdateError", func(t *testing.T) {
					t.Parallel()
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
				})
			})
		})

		suite.Run("Upsert", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Insert", func(t *testing.T) {
				t.Parallel()
				newID := 10
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := &TestItem{}
					return rw.Upsert(makeID(newID), testProto, func() error {
						if testProto.ID != "" {
							return errors.Errorf("Upsert of new row modified the ID, new value: %s", testProto.ID)
						}
						if testProto.Value != "" {
							return errors.Errorf("Upsert of new row modified the Value, new value: %s", testProto.Value)
						}
						testProto.ID = makeID(newID)
						testProto.Value = changedValue
						return nil
					})
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Added: []int{newID}})
			})

			subsuite.Run("ErrorInCallback", func(t *testing.T) {
				t.Parallel()
				newID := 10
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := &TestItem{}
					return rw.Upsert(makeID(newID), testProto, func() error {
						return &TestError{}
					})
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				overwriteID := 5
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					testProto := &TestItem{}
					return rw.Upsert(makeID(overwriteID), testProto, func() error {
						if testProto.ID != makeID(overwriteID) {
							return errors.Errorf("Upsert of existing row does not pass through the ID, expected: %s, actual: %s", makeID(overwriteID), testProto.ID)
						}
						if testProto.Value != originalValue {
							return errors.Errorf("Upsert of existing row does not pass through the value, expected: %s, actual: %s", originalValue, testProto.Value)
						}
						testProto.Value = changedValue
						return nil
					})
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Changed: []int{overwriteID}})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()
				subsuite.Run("UpsertError", func(t *testing.T) {
					t.Parallel()
					notExistsID := 10
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						if err := rw.Delete(makeID(6)); err != nil {
							return err
						}
						return rw.Delete(makeID(notExistsID))
					})
					require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: makeID(notExistsID)}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &TestItem{}
						if err := rw.Upsert(makeID(6), testProto, func() error {
							return nil
						}); err != nil {
							return err
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserErrorInCallback", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &TestItem{}
						if err := rw.Upsert(makeID(5), testProto, func() error {
							testProto.Value = changedValue
							return nil
						}); err != nil {
							return err
						}

						return rw.Upsert(makeID(6), testProto, func() error {
							return &TestError{}
						})
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("Delete", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				deleteID := 3
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					return rw.Delete(makeID(deleteID))
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Deleted: []int{deleteID}})
			})

			subsuite.Run("NotExists", func(t *testing.T) {
				t.Parallel()
				notExistsID := 10
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					return rw.Delete(makeID(notExistsID))
				})
				require.YesError(t, err)
				require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
				require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: makeID(notExistsID)}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()
				subsuite.Run("DeleteError", func(t *testing.T) {
					t.Parallel()
					notExistsID := 10
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						if err := rw.Delete(makeID(6)); err != nil {
							return err
						}
						return rw.Delete(makeID(notExistsID))
					})
					require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: makeID(notExistsID)}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						if err := rw.Delete(makeID(6)); err != nil {
							return err
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("DeleteAll", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				readOnly, writer := initCollection(t)
				err := writer(func(rw col.ReadWriteCollection) error {
					return rw.DeleteAll()
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Deleted: intRange(0, 10)})
				count, err := readOnly.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})
		})
	})
}

// test watches
// test indexes
// test multiple changes to the same row in one txn
// test interruption
