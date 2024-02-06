package bbolt

import (
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
