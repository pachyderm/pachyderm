package collection_test

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

const (
	defaultCollectionSize = 10
	originalValue         = "old"
	changedValue          = "new"
	collectionName        = "test_items"
)

var (
	TestSecondaryIndex = &col.Index{
		Name: "Value",
		Extract: func(val proto.Message) string {
			return val.(*col.TestItem).Value
		},
	}
)

type TestError struct{}

func (te TestError) Is(other error) bool {
	_, ok := other.(TestError)
	return ok
}

func (te TestError) Error() string {
	return "TestError"
}

type ReadCallback func(context.Context) col.ReadOnlyCollection
type WriteCallback func(context.Context, func(col.ReadWriteCollection) error) error

func idRange(start int, end int) []string {
	result := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, makeID(i))
	}
	return result
}

func putItem(item *col.TestItem) func(rw col.ReadWriteCollection) error {
	return func(rw col.ReadWriteCollection) error {
		return errors.EnsureStack(rw.Put(item.Id, item))
	}
}

// canceledContext is a helper function to provide a context that is already canceled
func canceledContext() context.Context {
	ctx, cancel := pctx.WithCancel(pctx.Child(pctx.TODO(), "canceled"))
	cancel()
	return ctx
}

// populateCollection writes a standard set of rows to the collection for use by
// individual tests.
func populateCollection(rw col.ReadWriteCollection) error {
	for _, id := range idRange(0, defaultCollectionSize) {
		if err := putItem(makeProto(id, originalValue))(rw); err != nil {
			return err
		}
	}
	return nil
}

func initCollection(
	ctx context.Context,
	t *testing.T,
	newCollection func(context.Context, *testing.T, ...bool) (ReadCallback, WriteCallback),
) (col.ReadOnlyCollection, WriteCallback) {
	reader, writer := newCollection(pctx.Child(ctx, "collection"), t)
	require.NoError(t, writer(pctx.Child(ctx, "writer"), populateCollection))
	return reader(pctx.Child(ctx, "reader")), writer
}

// Helper function to turn an int ID into a string so we don't need to use string literals
func makeID(i int) string {
	return fmt.Sprintf("%d", i)
}

// Helper function to instantiate the proto for a row by ID
func makeProto(id string, value ...string) *col.TestItem {
	if len(value) > 0 {
		return &col.TestItem{Id: id, Value: value[0]}
	}
	return &col.TestItem{Id: id, Value: changedValue}
}

type RowDiff struct {
	Deleted []string
	Changed []string
	Created []string
}

// checkDefaultCollection validates that the contents of the collection match
// the expected set of rows based on a diff from the default collection
// populated by populateCollection.
func checkDefaultCollection(t *testing.T, ro col.ReadOnlyCollection, diff RowDiff) {
	expected := map[string]string{}
	for _, id := range idRange(0, defaultCollectionSize) {
		expected[id] = originalValue
	}
	if diff.Deleted != nil {
		for _, id := range diff.Deleted {
			_, ok := expected[id]
			require.True(t, ok, "test specified a deleted row that was not in the original set")
			delete(expected, id)
		}
	}
	if diff.Changed != nil {
		for _, id := range diff.Changed {
			_, ok := expected[id]
			require.True(t, ok, "test specified a changed row that was not in the original set")
			expected[id] = changedValue
		}
	}
	if diff.Created != nil {
		for _, id := range diff.Created {
			_, ok := expected[id]
			require.False(t, ok, "test specified an added row that was already in the original set")
			expected[id] = changedValue
		}
	}
	checkCollection(t, ro, expected)
}

func checkItem(t *testing.T, item *col.TestItem, id string, value string) error {
	if item.Id != id {
		return errors.Errorf("Item has incorrect ID, expected: %s, actual: %s", id, item.Id)
	}
	if item.Value != value {
		return errors.Errorf("Item has incorrect Value, expected: %s, actual: %s", value, item.Value)
	}
	return nil
}

func checkCollection(t *testing.T, ro col.ReadOnlyCollection, expected map[string]string) {
	testProto := &col.TestItem{}
	actual := map[string]string{}
	require.NoError(t, ro.List(testProto, col.DefaultOptions(), func(id string) error {
		require.Equal(t, id, testProto.Id)
		actual[testProto.Id] = testProto.Value
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

type TestConstructor = func(context.Context, *testing.T) (ReadCallback, WriteCallback)

func collectionTests(
	ctx context.Context,
	parent *testing.T,
	newCollection func(context.Context, *testing.T, ...bool) (ReadCallback, WriteCallback),
) {
	parent.Run("ReadOnly", func(suite *testing.T) {
		suite.Parallel()
		emptyReader, _ := newCollection(ctx, suite)
		emptyRead := emptyReader(ctx)
		defaultRead, _ := initCollection(ctx, suite, newCollection)

		suite.Run("Get", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("ErrNotFound", func(t *testing.T) {
				t.Parallel()
				testProto := &col.TestItem{}
				err := defaultRead.Get("baz", testProto)
				require.True(t, errors.Is(err, col.ErrNotFound{}), "Incorrect error: %v", err)
				require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
			})

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				testProto := &col.TestItem{}
				require.NoError(t, defaultRead.Get("5", testProto))
				require.Equal(t, "5", testProto.Id)
			})
		})

		suite.Run("GetByIndex", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("Empty", func(t *testing.T) {
				t.Parallel()
				testProto := &col.TestItem{}
				err := emptyRead.GetByIndex(TestSecondaryIndex, "foo", testProto, col.DefaultOptions(), func(key string) error {
					return errors.New("GetByIndex callback should not have been called for an empty collection")
				})
				require.NoError(t, err)
			})

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				testProto := &col.TestItem{}
				keys := []string{}
				err := defaultRead.GetByIndex(TestSecondaryIndex, originalValue, testProto, col.DefaultOptions(), func(key string) error {
					require.Equal(t, testProto.Id, key)
					require.Equal(t, testProto.Value, originalValue)
					keys = append(keys, key)
					// Clear testProto.ID and testProto.Value just to make sure they get overwritten each time
					testProto.Id = ""
					testProto.Value = ""
					return nil
				})
				require.NoError(t, err)
				require.ElementsEqual(t, keys, idRange(0, defaultCollectionSize))
			})

			subsuite.Run("NoResults", func(t *testing.T) {
				t.Parallel()
				testProto := &col.TestItem{}
				err := defaultRead.GetByIndex(TestSecondaryIndex, changedValue, testProto, col.DefaultOptions(), func(string) error {
					return errors.New("GetByIndex callback should not have been called for an index value with no rows")
				})
				require.NoError(t, err)
			})

			subsuite.Run("InvalidIndex", func(t *testing.T) {
				t.Parallel()
				t.Skip("etcd collections do not validate their indexes")
				err := defaultRead.GetByIndex(&col.Index{}, "", &col.TestItem{}, col.DefaultOptions(), func(key string) error {
					return errors.New("GetByIndex callback should not have been called when using an invalid index")
				})
				require.YesError(t, err)
				require.Matches(t, "Unknown collection index", err.Error())
			})

			subsuite.Run("Partitioned", func(t *testing.T) {
				t.Parallel()
				partitionedReader, partitionedWriter := newCollection(ctx, suite)
				partitionedRead := partitionedReader(ctx)
				require.NoError(t, partitionedWriter(ctx, populateCollection))

				expected := map[string][]string{
					"a": idRange(0, 3),
					"b": idRange(3, 6),
					"c": idRange(6, 9),
					"d": idRange(9, 12),
				}

				require.NoError(t, partitionedWriter(ctx, func(rw col.ReadWriteCollection) error {
					for idxVal, ids := range expected {
						for _, id := range ids {
							testProto := &col.TestItem{Id: id, Value: idxVal}
							if err := rw.Put(id, testProto); err != nil {
								return errors.EnsureStack(err)
							}
						}
					}
					return nil
				}))

				for idxVal, ids := range expected {
					testProto := &col.TestItem{}
					keys := []string{}
					err := partitionedRead.GetByIndex(TestSecondaryIndex, idxVal, testProto, col.DefaultOptions(), func(key string) error {
						require.Equal(t, testProto.Id, key)
						require.Equal(t, testProto.Value, idxVal)
						keys = append(keys, key)
						// Clear testProto.ID and testProto.Value just to make sure they get overwritten each time
						testProto.Id = ""
						testProto.Value = ""
						return nil
					})
					require.NoError(t, err)
					require.ElementsEqual(t, keys, ids)
				}
			})
		})

		suite.Run("List", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("Empty", func(t *testing.T) {
				t.Parallel()
				testProto := &col.TestItem{}
				err := emptyRead.List(testProto, col.DefaultOptions(), func(string) error {
					return errors.New("List callback should not have been called for an empty collection")
				})
				require.NoError(t, err)
			})

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				items := map[string]string{}
				testProto := &col.TestItem{}
				err := defaultRead.List(testProto, col.DefaultOptions(), func(key string) error {
					require.Equal(t, testProto.Id, key)
					items[testProto.Id] = testProto.Value
					// Clear testProto.ID and testProto.Value just to make sure they get overwritten each time
					testProto.Id = ""
					testProto.Value = ""
					return nil
				})
				require.NoError(t, err)

				keys := []string{}
				for k, v := range items {
					keys = append(keys, k)
					require.Equal(t, originalValue, v)
				}
				expectedKeys := idRange(0, defaultCollectionSize)
				require.ElementsEqual(t, expectedKeys, keys)
			})

			subsuite.Run("Sort", func(subsuite *testing.T) {
				subsuite.Parallel()

				testSort := func(t *testing.T, ro col.ReadOnlyCollection, target col.SortTarget, expectedAsc []string) {
					t.Parallel()

					collect := func(order col.SortOrder) []string {
						keys := []string{}
						testProto := &col.TestItem{}
						err := ro.List(testProto, &col.Options{Target: target, Order: order}, func(string) error {
							keys = append(keys, testProto.Id)
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
					// The default collection was written in a single transaction, so all
					// the rows have the same created time - make our own
					createReader, writer := newCollection(ctx, t)
					createCol := createReader(ctx)

					createKeys := []string{"0", "6", "7", "9", "3", "8", "4", "1", "2", "5"}
					for _, k := range createKeys {
						err := writer(ctx, func(rw col.ReadWriteCollection) error {
							testProto := &col.TestItem{Id: k, Value: originalValue}
							return errors.EnsureStack(rw.Create(k, testProto))
						})
						require.NoError(t, err)
					}

					checkDefaultCollection(t, createCol, RowDiff{})
					testSort(t, createCol, col.SortByCreateRevision, createKeys)
				})

				subsuite.Run("UpdatedAt", func(t *testing.T) {
					// Create a new collection that we can modify to get a different ordering here
					reader, writer := newCollection(ctx, t, true)
					modRead := reader(ctx)
					require.NoError(suite, writer(ctx, populateCollection))

					modKeys := []string{"5", "7", "2", "9", "1", "0", "8", "4", "3", "6"}
					for _, k := range modKeys {
						err := writer(ctx, func(rw col.ReadWriteCollection) error {
							testProto := &col.TestItem{}
							if err := rw.Update(k, testProto, func() error {
								testProto.Value = changedValue
								return nil
							}); err != nil {
								return errors.EnsureStack(err)
							}
							return nil
						})
						require.NoError(t, err)
					}
					testSort(t, modRead, col.SortByModRevision, modKeys)
				})

				subsuite.Run("Key", func(t *testing.T) {
					testSort(t, defaultRead, col.SortByKey, idRange(0, 10))
				})
			})

			subsuite.Run("ErrorInCallback", func(t *testing.T) {
				t.Parallel()
				count := 0
				testProto := &col.TestItem{}
				err := defaultRead.List(testProto, col.DefaultOptions(), func(string) error {
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
				testProto := &col.TestItem{}
				err := defaultRead.List(testProto, col.DefaultOptions(), func(string) error {
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
				count, err := defaultRead.Count()
				require.NoError(t, err)
				require.Equal(t, int64(10), count)

				count, err = emptyRead.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})
		})
	})

	parent.Run("ReadWrite", func(suite *testing.T) {
		suite.Parallel()
		testRollback := func(t *testing.T, cb func(rw col.ReadWriteCollection) error) error {
			readOnly, writer := initCollection(ctx, t, newCollection)
			err := writer(ctx, func(rw col.ReadWriteCollection) error {
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
				readOnly, writer := initCollection(ctx, t, newCollection)

				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					if err := rw.Get(makeID(8), testProto); err != nil {
						return errors.EnsureStack(err)
					}
					if testProto.Value != originalValue {
						return errors.Errorf("Incorrect value when fetching via ReadWriteCollection.Get, expected: %s, actual: %s", originalValue, testProto.Value)
					}
					return nil
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("AfterDelete", func(t *testing.T) {
				t.Parallel()
				readOnly, writer := initCollection(ctx, t, newCollection)

				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					if err := rw.Get(makeID(1), testProto); err != nil {
						return errors.EnsureStack(err)
					}

					oobID := makeID(10)
					if err := rw.Get(oobID, testProto); !col.IsErrNotFound(err) {
						return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", oobID)
					}

					require.NoError(t, rw.DeleteAll())

					for _, id := range idRange(0, defaultCollectionSize) {
						if err := rw.Get(id, testProto); !col.IsErrNotFound(err) {
							return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", id)
						}
					}
					return nil
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Deleted: idRange(0, defaultCollectionSize)})

				count, err := readOnly.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})
		})
		suite.Run("Create", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				newID := makeID(10)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := makeProto(newID)
					return errors.EnsureStack(rw.Create(testProto.Id, testProto))
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Created: []string{newID}})
			})

			subsuite.Run("ErrExists", func(t *testing.T) {
				t.Parallel()
				overwriteID := makeID(5)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := makeProto(overwriteID)
					return errors.EnsureStack(rw.Create(testProto.Id, testProto))
				})
				require.YesError(t, err)
				require.True(t, col.IsErrExists(err), "Incorrect error: %v", err)
				require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: overwriteID}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()
				subsuite.Run("CreateError", func(t *testing.T) {
					t.Parallel()
					newID := makeID(10)
					overwriteID := makeID(6)
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(newID)
						if err := rw.Create(testProto.Id, testProto); err != nil {
							return errors.EnsureStack(err)
						}
						testProto = makeProto(overwriteID)
						return errors.EnsureStack(rw.Create(testProto.Id, testProto))
					})
					require.True(t, col.IsErrExists(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: overwriteID}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(makeID(10))
						if err := rw.Create(testProto.Id, testProto); err != nil {
							return errors.EnsureStack(err)
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})

				subsuite.Run("NoRollbackOnExists", func(t *testing.T) {
					t.Parallel()
					overwriteID := makeID(5)
					newID := makeID(10)
					readOnly, writer := initCollection(ctx, t, newCollection)
					err := writer(ctx, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(overwriteID)
						if err := rw.Create(testProto.Id, testProto); err == nil {
							return errors.New("Expected col.ErrExists, but got no error")
						} else if !col.IsErrExists(err) {
							return errors.EnsureStack(err)
						}
						testProto = makeProto(newID)
						return errors.EnsureStack(rw.Create(testProto.Id, testProto))
					})
					require.NoError(t, err)
					checkDefaultCollection(t, readOnly, RowDiff{Created: []string{newID}})
				})
			})
		})

		suite.Run("Put", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Insert", func(t *testing.T) {
				t.Parallel()
				newID := makeID(10)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := makeProto(newID)
					return errors.EnsureStack(rw.Put(testProto.Id, testProto))
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Created: []string{newID}})
			})

			subsuite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				overwriteID := makeID(5)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := makeProto(overwriteID)
					return errors.EnsureStack(rw.Put(testProto.Id, testProto))
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Changed: []string{overwriteID}})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()
				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(makeID(10))
						if err := rw.Put(testProto.Id, testProto); err != nil {
							return errors.EnsureStack(err)
						}
						testProto = makeProto(makeID(8))
						if err := rw.Put(testProto.Id, testProto); err != nil {
							return errors.EnsureStack(err)
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
				updateID := makeID(1)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					if err := rw.Upsert(updateID, testProto, func() error {
						if err := checkItem(t, testProto, updateID, originalValue); err != nil {
							return err
						}
						testProto.Value = changedValue
						return nil
					}); err != nil {
						return errors.EnsureStack(err)
					}
					if err := rw.Get(updateID, testProto); err != nil {
						return errors.EnsureStack(err)
					}
					return checkItem(t, testProto, updateID, changedValue)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Changed: []string{updateID}})
			})

			subsuite.Run("ErrorInCallback", func(t *testing.T) {
				t.Parallel()
				updateID := makeID(2)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					return errors.EnsureStack(rw.Update(updateID, testProto, func() error {
						return &TestError{}
					}))
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("NotFound", func(t *testing.T) {
				t.Parallel()
				notExistsID := makeID(10)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					return errors.EnsureStack(rw.Update(notExistsID, testProto, func() error {
						return nil
					}))
				})
				require.YesError(t, err)
				require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
				require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: notExistsID}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()

				subsuite.Run("ErrorInCallback", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Update(makeID(2), testProto, func() error {
							testProto.Value = changedValue
							return nil
						}); err != nil {
							return errors.EnsureStack(err)
						}
						return errors.EnsureStack(rw.Update(makeID(2), testProto, func() error {
							return &TestError{}
						}))
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})

				subsuite.Run("UpdateError", func(t *testing.T) {
					t.Parallel()
					notExistsID := makeID(10)
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Update(makeID(3), testProto, func() error {
							testProto.Value = changedValue
							return nil
						}); err != nil {
							return errors.EnsureStack(err)
						}
						return errors.EnsureStack(rw.Update(notExistsID, testProto, func() error {
							testProto.Value = changedValue
							return nil
						}))
					})
					require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: notExistsID}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Update(makeID(6), testProto, func() error {
							testProto.Value = changedValue
							return nil
						}); err != nil {
							return errors.EnsureStack(err)
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("Upsert", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("Insert", func(t *testing.T) {
				t.Parallel()
				newID := makeID(10)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					return errors.EnsureStack(rw.Upsert(newID, testProto, func() error {
						if testProto.Id != "" {
							return errors.Errorf("Upsert of new row modified the ID, new value: %s", testProto.Id)
						}
						if testProto.Value != "" {
							return errors.Errorf("Upsert of new row modified the Value, new value: %s", testProto.Value)
						}
						testProto.Id = newID
						testProto.Value = changedValue
						return nil
					}))
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Created: []string{newID}})
			})

			subsuite.Run("ErrorInCallback", func(t *testing.T) {
				t.Parallel()
				newID := makeID(10)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					return errors.EnsureStack(rw.Upsert(newID, testProto, func() error {
						return &TestError{}
					}))
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				overwriteID := makeID(5)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					err := rw.Upsert(overwriteID, testProto, func() error {
						if testProto.Id != overwriteID {
							return errors.Errorf("Upsert of existing row does not pass through the ID, expected: %s, actual: %s", overwriteID, testProto.Id)
						}
						if testProto.Value != originalValue {
							return errors.Errorf("Upsert of existing row does not pass through the value, expected: %s, actual: %s", originalValue, testProto.Value)
						}
						testProto.Value = changedValue
						return nil
					})
					return errors.EnsureStack(err)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Changed: []string{overwriteID}})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()

				subsuite.Run("UpsertError", func(t *testing.T) {
					t.Parallel()
					existsID := makeID(3)
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Upsert(makeID(6), testProto, func() error {
							return nil
						}); err != nil {
							return errors.EnsureStack(err)
						}
						return errors.EnsureStack(rw.Create(existsID, testProto))
					})
					require.True(t, col.IsErrExists(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: existsID}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Upsert(makeID(6), testProto, func() error {
							return nil
						}); err != nil {
							return errors.EnsureStack(err)
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserErrorInCallback", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Upsert(makeID(5), testProto, func() error {
							testProto.Value = changedValue
							return nil
						}); err != nil {
							return errors.EnsureStack(err)
						}

						return errors.EnsureStack(rw.Upsert(makeID(6), testProto, func() error {
							return &TestError{}
						}))
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("Delete", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				deleteID := makeID(3)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					return errors.EnsureStack(rw.Delete(deleteID))
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Deleted: []string{deleteID}})
			})

			subsuite.Run("NotExists", func(t *testing.T) {
				t.Parallel()
				notExistsID := makeID(10)
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					return errors.EnsureStack(rw.Delete(notExistsID))
				})
				require.YesError(t, err)
				require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
				require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: notExistsID}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()

				subsuite.Run("DeleteError", func(t *testing.T) {
					t.Parallel()
					notExistsID := makeID(10)
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						if err := rw.Delete(makeID(6)); err != nil {
							return errors.EnsureStack(err)
						}
						return errors.EnsureStack(rw.Delete(notExistsID))
					})
					require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: notExistsID}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						if err := rw.Delete(makeID(6)); err != nil {
							return errors.EnsureStack(err)
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
				readOnly, writer := initCollection(ctx, t, newCollection)
				err := writer(ctx, func(rw col.ReadWriteCollection) error {
					return errors.EnsureStack(rw.DeleteAll())
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Deleted: idRange(0, 10)})
				count, err := readOnly.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})

			subsuite.Run("TransactionRollback", func(t *testing.T) {
				t.Parallel()
				err := testRollback(t, func(rw col.ReadWriteCollection) error {
					if err := rw.DeleteAll(); err != nil {
						return errors.EnsureStack(err)
					}
					return &TestError{}
				})
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
			})
		})
	})
}

// TODO: test indexes
// TODO: test multiple changes to the same row in one txn
// TODO: test interruption
