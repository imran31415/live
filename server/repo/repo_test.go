package repo_test

import (
	"admin/models"
	"admin/server/repo"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"testing"
)

func TestSqlRepo_UpdateUserBySubId(t *testing.T) {
	//todo
}

// TestSqlRepo_GetUserById checks that our Repo layer can successfully retrieve records for users by ID
func TestSqlRepo_GetUserById(t *testing.T) {
	tests := []struct {
		Name      string
		In        *models.User
		Exp       *models.User
		IsErrored bool
	}{
		{
			Name: "Get a User",
			In: &models.User{
				Model: gorm.Model{
					ID: 100,
				},
				FirstName: "foo",
				LastName:  "bar",
				Email:     "nelson@nelson.com",
			},
			Exp: &models.User{
				Model: gorm.Model{
					ID: 100,
				},
				FirstName: "foo",
				LastName:  "bar",
				Email:     "nelson@nelson.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			// Connect to local sqlite database for testing
			name := fmt.Sprintf("./%d.db", rand.Int())
			// Cleanup test database at end of test

			r := setupTestDb(t, name)
			defer os.Remove(name)

			// Clean up test data at the end
			// Note: we use Unscoped delete to remove the record permanently
			//    - Not using unscoped means the record will still exist and occupy the primary key, even though its marked as deleted
			defer r.DB.Unscoped().Delete(tt.In)

			// Delete user if it exists so we have a clean test
			err := r.DB.Unscoped().Delete(tt.In).Error
			require.NoError(t, err)

			require.NoError(t, r.DB.Create(tt.In).Error)

			// Test we are able to load a user, and it matches what we expect
			user, dbErr := r.GetUserById(tt.In.ID)
			require.NoError(t, dbErr)
			// check that fields match that we care about
			// Note we don't check the objects match directly since GORM adds some sugar to the fields such as last modified, etc.

			require.Equal(t, tt.Exp.ID, user.ID)
			require.Equal(t, tt.Exp.FirstName, user.FirstName)
			require.Equal(t, tt.Exp.LastName, user.LastName)
			require.Equal(t, tt.Exp.Email, user.Email)

		})
	}
}

func setupTestDb(t *testing.T, name string) *repo.SqlRepo {
	db, err := gorm.Open("sqlite3", name)
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}).Error
	require.NoError(t, err)

	return &repo.SqlRepo{DB: db}
}
