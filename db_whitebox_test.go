package bbolt

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMethodPage(t *testing.T) {
	testCases := []struct {
		name          string
		readonly      bool
		expectedError error
	}{
		{
			name:          "write mode",
			readonly:      false,
			expectedError: nil,
		},
		{
			name:          "readonly mode with preloading free pages",
			readonly:      true,
			expectedError: nil,
		},
	}

	fileName, err := prepareData(t)
	require.NoError(t, err)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db, err := Open(fileName, 0666, &Options{
				ReadOnly: tc.readonly,
			})
			require.NoError(t, err)
			defer db.Close()

			tx, err := db.Begin(!tc.readonly)
			require.NoError(t, err)

			_, err = tx.Page(0)
			require.Equal(t, tc.expectedError, err)

			if tc.readonly {
				require.NoError(t, tx.Rollback())
			} else {
				require.NoError(t, tx.Commit())
			}

			require.NoError(t, db.Close())
		})
	}
}

// Ensure that a recovered panic in Update, while a reader keeps every freed
// page pending, does not put pending pages on the free list. Such pages are
// otherwise reallocated while still referenced, corrupting the B+tree.
func TestDB_Update_PanicRollbackKeepsPendingPagesUsed(t *testing.T) {
	for _, typ := range []FreelistType{FreelistArrayType, FreelistMapType} {
		t.Run(string(typ), func(t *testing.T) {
			// A large initial mmap avoids remapping, which would block on the open reader.
			db, err := Open(filepath.Join(t.TempDir(), "db"), 0600, &Options{
				FreelistType:    typ,
				InitialMmapSize: 1 << 28,
			})
			require.NoError(t, err)
			defer db.Close()

			put := func(lo, hi int) {
				err := db.Update(func(tx *Tx) error {
					b, err := tx.CreateBucketIfNotExists([]byte("b"))
					if err != nil {
						return err
					}
					for i := lo; i < hi; i++ {
						if err := b.Put([]byte(fmt.Sprintf("k%06d", i)), make([]byte, 500)); err != nil {
							return err
						}
					}
					return nil
				})
				require.NoError(t, err)
			}
			put(0, 2000)

			reader, err := db.Begin(false)
			require.NoError(t, err)
			// Release the reader before db.Close on failure; Close waits for open readers.
			defer func() { _ = reader.Rollback() }()
			for db.freelist.free_count() > 0 {
				put(0, 2000)
			}
			require.NotZero(t, db.freelist.pending_count())

			func() {
				defer func() { require.NotNil(t, recover()) }()
				_ = db.Update(func(tx *Tx) error { panic("rollback") })
			}()

			pending := make(map[pgid]bool)
			for _, txp := range db.freelist.pending {
				for _, id := range txp.ids {
					pending[id] = true
				}
			}
			for _, id := range db.freelist.getFreePageIDs() {
				require.Falsef(t, pending[id], "page %d is both free and pending", id)
			}

			put(0, 1000)
			require.NoError(t, reader.Rollback())
			for i := 0; i < 10; i++ {
				put(i*200, i*200+300)
			}

			require.NoError(t, db.View(func(tx *Tx) error {
				var errs []error
				for err := range tx.Check() {
					errs = append(errs, err)
				}
				require.Empty(t, errs)
				return nil
			}))
		})
	}
}

func prepareData(t *testing.T) (string, error) {
	fileName := filepath.Join(t.TempDir(), "db")
	db, err := Open(fileName, 0666, nil)
	if err != nil {
		return "", err
	}
	if err := db.Close(); err != nil {
		return "", err
	}

	return fileName, nil
}
