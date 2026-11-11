package erasure

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStores struct {
	calls []string

	keys      []string
	failRows  error
	failS3    map[string]error
	failMongo error
	failKeys  error

	deletedObjects []string
}

func (f *fakeStores) PhotoObjectKeys(context.Context, int64) ([]string, error) {
	f.calls = append(f.calls, "keys")

	return f.keys, f.failKeys
}

func (f *fakeStores) DeleteUserRows(context.Context, int64) error {
	f.calls = append(f.calls, "rows")

	return f.failRows
}

func (f *fakeStores) Delete(_ context.Context, key string) error {
	f.calls = append(f.calls, "s3:"+key)

	if err := f.failS3[key]; err != nil {
		return err
	}

	f.deletedObjects = append(f.deletedObjects, key)

	return nil
}

func (f *fakeStores) DeleteByUserID(context.Context, int64) error {
	f.calls = append(f.calls, "mongo")

	return f.failMongo
}

func newService(f *fakeStores) Service { return NewService(f, f, f) }

func TestDeleteUserData_ClearsExternalStoresBeforeRows(t *testing.T) {
	f := &fakeStores{keys: []string{"a.jpg", "b.jpg"}}

	require.NoError(t, newService(f).DeleteUserData(context.Background(), 7))

	assert.Equal(t, []string{"keys", "s3:a.jpg", "s3:b.jpg", "mongo", "rows"}, f.calls,
		"rows hold the object keys, so they must be removed last")
}

func TestDeleteUserData_WithNoDataIsANoOpSuccess(t *testing.T) {
	f := &fakeStores{}

	require.NoError(t, newService(f).DeleteUserData(context.Background(), 7))
	assert.Equal(t, []string{"keys", "mongo", "rows"}, f.calls)
}

func TestDeleteUserData_FailureStopsAndKeepsRowsForRetry(t *testing.T) {
	boom := errors.New("boom")

	cases := map[string]struct {
		stores    *fakeStores
		wantCalls []string
	}{
		"s3 fails": {
			stores:    &fakeStores{keys: []string{"a.jpg"}, failS3: map[string]error{"a.jpg": boom}},
			wantCalls: []string{"keys", "s3:a.jpg"},
		},
		"mongo fails": {
			stores:    &fakeStores{keys: []string{"a.jpg"}, failMongo: boom},
			wantCalls: []string{"keys", "s3:a.jpg", "mongo"},
		},
		"rows fail": {
			stores:    &fakeStores{failRows: boom},
			wantCalls: []string{"keys", "mongo", "rows"},
		},
		"key lookup fails": {
			stores:    &fakeStores{failKeys: boom},
			wantCalls: []string{"keys"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := newService(tc.stores).DeleteUserData(context.Background(), 7)

			require.ErrorIs(t, err, boom, "a failure must never be reported as success")
			assert.Equal(t, tc.wantCalls, tc.stores.calls)
		})
	}
}

func TestDeleteUserData_RetryAfterPartialFailureCompletes(t *testing.T) {
	boom := errors.New("s3 timeout")
	f := &fakeStores{keys: []string{"a.jpg", "b.jpg"}, failS3: map[string]error{"b.jpg": boom}}
	svc := newService(f)

	require.ErrorIs(t, svc.DeleteUserData(context.Background(), 7), boom)
	assert.Equal(t, []string{"a.jpg"}, f.deletedObjects)
	assert.NotContains(t, f.calls, "rows", "rows must survive so the retry still knows the keys")

	f.failS3 = nil
	f.calls = nil

	require.NoError(t, svc.DeleteUserData(context.Background(), 7))
	assert.Equal(t, "rows", f.calls[len(f.calls)-1], "rows are removed only after everything else succeeded")
	assert.Contains(t, f.deletedObjects, "b.jpg", "the object that failed is deleted on retry")
	assert.Contains(t, f.deletedObjects, "a.jpg", "deleting an already deleted object again is harmless")
}
