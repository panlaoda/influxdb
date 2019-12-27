package kv_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/kv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	newStoreBase := func(t *testing.T, bktSuffix string, encKeyFn, encBodyFn kv.EncodeEntFn, decFn kv.DecodeBucketEntFn, decToEntFn kv.DecodedValToEntFn) (*kv.StoreBase, kv.Store) {
		t.Helper()

		inmemSVC, _, _ := NewTestInmemStore(t)
		store := kv.NewStoreBase("foo", []byte("foo_"+bktSuffix), encKeyFn, encBodyFn, decFn, decToEntFn)

		require.NoError(t, inmemSVC.Update(context.TODO(), func(tx kv.Tx) error {
			return store.Init(tx)
		}))
		return store, inmemSVC
	}

	newFooStoreBase := func(t *testing.T, bktSuffix string) (*kv.StoreBase, kv.Store) {
		return newStoreBase(t, bktSuffix, kv.EncIDKey, kv.EncBodyJSON, decJSONFooFn, decFooEntFn)
	}

	t.Run("Put", func(t *testing.T) {
		base, inmemStore := newFooStoreBase(t, "put")
		testPutBase(t, inmemStore, base, base.BktName)
	})

	t.Run("DeleteEnt", func(t *testing.T) {
		base, inmemStore := newFooStoreBase(t, "delete_ent")

		testDeleteEntBase(t, inmemStore, base)
	})

	t.Run("Delete", func(t *testing.T) {
		testDeleteBase(t, func(t *testing.T, suffix string) (storeBase, kv.Store) {
			return newFooStoreBase(t, suffix)
		})
	})

	t.Run("FindEnt", func(t *testing.T) {
		base, inmemStore := newFooStoreBase(t, "find_ent")
		testFindEnt(t, inmemStore, base)
	})

	t.Run("Find", func(t *testing.T) {
		testFind(t, func(t *testing.T, suffix string) (storeBase, kv.Store) {
			return newFooStoreBase(t, suffix)
		})
	})
}

func testPutBase(t *testing.T, kvStore kv.Store, base storeBase, bktName []byte) foo {
	t.Helper()

	expected := foo{
		ID:    1,
		OrgID: 9000,
		Name:  "foo_1",
	}

	update(t, kvStore, func(tx kv.Tx) error {
		return base.Put(tx, kv.Entity{
			ID:    expected.ID,
			Name:  expected.Name,
			OrgID: expected.OrgID,
			Body:  expected,
		})
	})

	var actual foo
	decodeJSON(t, getEntRaw(t, kvStore, bktName, encodeID(t, expected.ID)), &actual)

	assert.Equal(t, expected, actual)

	return expected
}

func testDeleteEntBase(t *testing.T, kvStore kv.Store, base storeBase) kv.Entity {
	t.Helper()

	expected := newFooEnt(1, 9000, "foo_1")
	seedEnts(t, kvStore, base, expected)

	update(t, kvStore, func(tx kv.Tx) error {
		return base.DeleteEnt(tx, kv.Entity{ID: expected.ID})
	})

	err := kvStore.View(context.TODO(), func(tx kv.Tx) error {
		_, err := base.FindEnt(tx, kv.Entity{ID: expected.ID})
		return err
	})
	isNotFoundErr(t, err)
	return expected
}

func testDeleteBase(t *testing.T, fn func(t *testing.T, suffix string) (storeBase, kv.Store), assertFns ...func(*testing.T, kv.Store, storeBase, []foo)) {
	expectedEnts := []kv.Entity{
		newFooEnt(1, 9000, "foo_0"),
		newFooEnt(2, 9000, "foo_1"),
		newFooEnt(3, 9003, "foo_2"),
		newFooEnt(4, 9004, "foo_3"),
	}

	tests := []struct {
		name     string
		opts     kv.DeleteOpts
		expected []interface{}
	}{
		{
			name: "delete all",
			opts: kv.DeleteOpts{
				FilterFn: func(k []byte, v interface{}) bool {
					return true
				},
			},
		},
		{
			name: "delete IDs less than 4",
			opts: kv.DeleteOpts{
				FilterFn: func(k []byte, v interface{}) bool {
					if f, ok := v.(foo); ok {
						return f.ID < 4
					}
					return true
				},
			},
			expected: toIfaces(expectedEnts[3]),
		},
	}

	for _, tt := range tests {
		fn := func(t *testing.T) {
			t.Helper()

			base, inmemStore := fn(t, "delete")

			seedEnts(t, inmemStore, base, expectedEnts...)

			update(t, inmemStore, func(tx kv.Tx) error {
				return base.Delete(tx, tt.opts)
			})

			var actuals []interface{}
			view(t, inmemStore, func(tx kv.Tx) error {
				return base.Find(tx, kv.FindOpts{
					CaptureFn: func(key []byte, decodedVal interface{}) error {
						actuals = append(actuals, decodedVal)
						return nil
					},
				})
			})

			assert.Equal(t, tt.expected, actuals)

			var entsLeft []foo
			for _, expected := range tt.expected {
				ent, ok := expected.(foo)
				require.Truef(t, ok, "got: %#v", expected)
				entsLeft = append(entsLeft, ent)
			}

			for _, assertFn := range assertFns {
				assertFn(t, inmemStore, base, entsLeft)
			}
		}
		t.Run(tt.name, fn)
	}
}

func testFindEnt(t *testing.T, kvStore kv.Store, base storeBase) kv.Entity {
	t.Helper()

	expected := newFooEnt(1, 9000, "foo_1")
	seedEnts(t, kvStore, base, expected)

	var actual interface{}
	view(t, kvStore, func(tx kv.Tx) error {
		f, err := base.FindEnt(tx, kv.Entity{ID: expected.ID})
		actual = f
		return err
	})

	assert.Equal(t, expected.Body, actual)

	return expected
}

func testFind(t *testing.T, fn func(t *testing.T, suffix string) (storeBase, kv.Store)) {
	t.Helper()

	expectedEnts := []kv.Entity{
		newFooEnt(1, 9000, "foo_0"),
		newFooEnt(2, 9000, "foo_1"),
		newFooEnt(3, 9003, "foo_2"),
		newFooEnt(4, 9004, "foo_3"),
	}

	tests := []struct {
		name     string
		opts     kv.FindOpts
		expected []interface{}
	}{
		{
			name:     "no options",
			expected: toIfaces(expectedEnts...),
		},
		{
			name:     "with order descending",
			opts:     kv.FindOpts{Descending: true},
			expected: reverseSlc(toIfaces(expectedEnts...)),
		},
		{
			name:     "with limit",
			opts:     kv.FindOpts{Limit: 1},
			expected: toIfaces(expectedEnts[0]),
		},
		{
			name:     "with offset",
			opts:     kv.FindOpts{Offset: 1},
			expected: toIfaces(expectedEnts[1:]...),
		},
		{
			name: "with offset and limit",
			opts: kv.FindOpts{
				Limit:  1,
				Offset: 1,
			},
			expected: toIfaces(expectedEnts[1]),
		},
		{
			name: "with descending, offset, and limit",
			opts: kv.FindOpts{
				Descending: true,
				Limit:      1,
				Offset:     1,
			},
			expected: toIfaces(expectedEnts[2]),
		},
	}

	for _, tt := range tests {
		fn := func(t *testing.T) {
			base, kvStore := fn(t, "find")
			seedEnts(t, kvStore, base, expectedEnts...)

			var actuals []interface{}
			tt.opts.CaptureFn = func(key []byte, decodedVal interface{}) error {
				actuals = append(actuals, decodedVal)
				return nil
			}

			view(t, kvStore, func(tx kv.Tx) error {
				return base.Find(tx, tt.opts)
			})

			assert.Equal(t, tt.expected, actuals)
		}
		t.Run(tt.name, fn)
	}
}

type foo struct {
	ID    influxdb.ID
	OrgID influxdb.ID

	Name string
}

func decodeJSON(t *testing.T, b []byte, v interface{}) {
	t.Helper()
	require.NoError(t, json.Unmarshal(b, &v))
}

type storeBase interface {
	Delete(tx kv.Tx, opts kv.DeleteOpts) error
	DeleteEnt(tx kv.Tx, ent kv.Entity) error
	FindEnt(tx kv.Tx, ent kv.Entity) (interface{}, error)
	Find(tx kv.Tx, opts kv.FindOpts) error
	Put(tx kv.Tx, ent kv.Entity) error
}

func seedEnts(t *testing.T, kvStore kv.Store, store storeBase, ents ...kv.Entity) {
	t.Helper()

	for _, ent := range ents {
		update(t, kvStore, func(tx kv.Tx) error { return store.Put(tx, ent) })
	}
}

func update(t *testing.T, kvStore kv.Store, fn func(tx kv.Tx) error) {
	t.Helper()

	require.NoError(t, kvStore.Update(context.TODO(), fn))
}

func view(t *testing.T, kvStore kv.Store, fn func(tx kv.Tx) error) {
	t.Helper()
	require.NoError(t, kvStore.View(context.TODO(), fn))
}

func getEntRaw(t *testing.T, kvStore kv.Store, bktName []byte, key []byte) []byte {
	t.Helper()

	var actualRaw []byte
	err := kvStore.View(context.TODO(), func(tx kv.Tx) error {
		b, err := tx.Bucket(bktName)
		require.NoError(t, err)

		actualRaw, err = b.Get(key)
		return err
	})
	require.NoError(t, err)
	return actualRaw
}

func encodeID(t *testing.T, id influxdb.ID) []byte {
	t.Helper()

	b, err := id.Encode()
	require.NoError(t, err)
	return b
}

func decJSONFooFn(key, val []byte) ([]byte, interface{}, error) {
	var f foo
	if err := json.Unmarshal(val, &f); err != nil {
		return nil, nil, err
	}
	return key, f, nil
}

func decFooEntFn(k []byte, v interface{}) (kv.Entity, error) {
	f, ok := v.(foo)
	if !ok {
		return kv.Entity{}, fmt.Errorf("invalid entry: %#v", v)
	}
	return kv.Entity{
		ID:    f.ID,
		Name:  f.Name,
		OrgID: f.OrgID,
		Body:  f,
	}, nil
}

func newFooEnt(id, orgID influxdb.ID, name string) kv.Entity {
	f := foo{ID: id, Name: name, OrgID: orgID}
	return kv.Entity{
		ID:    f.ID,
		Name:  f.Name,
		OrgID: f.OrgID,
		Body:  f,
	}
}

func isNotFoundErr(t *testing.T, err error) {
	t.Helper()

	iErr, ok := err.(*influxdb.Error)
	if !ok {
		require.FailNowf(t, "expected an *influxdb.Error type", "got: %#v", err)
	}
	assert.Equal(t, influxdb.ENotFound, iErr.Code)
}

func toIfaces(ents ...kv.Entity) []interface{} {
	var actuals []interface{}
	for _, ent := range ents {
		actuals = append(actuals, ent.Body)
	}
	return actuals
}

func reverseSlc(slc []interface{}) []interface{} {
	for i, j := 0, len(slc)-1; i < j; i, j = i+1, j-1 {
		slc[i], slc[j] = slc[j], slc[i]
	}
	return slc
}
