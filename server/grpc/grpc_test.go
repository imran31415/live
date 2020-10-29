package grpc_test

import (
	"admin/models"
	pb "admin/protos"
	"admin/server/grpc"
	"admin/server/repo"
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"strings"
	"testing"
)

// TestServer_GetUserById checks that our grpc layer can successfully retrieve records for users by ID
func TestServer_GetUserById(t *testing.T) {
	tests := []struct {
		Name string
		In   *models.User
		Exp  *pb.User
	}{
		{
			Name: "Get a User",
			In: &models.User{
				Model: gorm.Model{
					ID: 100,
				},
				FirstName: "foo",
				LastName:  "bar",
			},
			Exp: &pb.User{
				Id:        100,
				FirstName: "foo",
				LastName:  "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			//============ Start Test Setup ============

			// Connect to local MYSQL database for testing
			name := fmt.Sprintf("./%d.db", rand.Int())
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
			//============ End Test Setup ============

			// ============Start Test ============

			serv := grpc.NewServerWithRepo(r, "", "")

			got, rErr := serv.GetUserById(context.Background(), &pb.Id{Id: int64(tt.In.ID)})
			require.NoError(t, rErr)
			require.Equal(t, tt.Exp, got)

		})
	}
}

// TestServer_GetUserById checks that our grpc layer can successfully retrieve records for users by ID
func TestServer_GetSignedImageUploadUrl(t *testing.T) {
	tests := []struct {
		Name string
		In   *pb.SignedImageUploadUrlRequest
		Exp  string
	}{
		{
			Name: "Get a signed URL",
			In:   &pb.SignedImageUploadUrlRequest{Ext: pb.SignedImageUploadUrlRequest_JPG},
			Exp:  "https://storage.googleapis.com/image-user-upload/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			name := fmt.Sprintf("./%d.db", rand.Int())
			r := setupTestDb(t, name)
			defer os.Remove(name)
			// Delete anything we will create before we start the test to avoid pollution

			// Test start
			serv := grpc.NewServerWithRepo(r, "", "")

			exp, err := serv.GetSignedImageUploadUrl(context.Background(), tt.In)
			require.NoError(t, err)
			require.True(t, strings.HasPrefix(exp.GetUrl(), tt.Exp))

		})
	}
}

func setupTestDb(t *testing.T, name string) *repo.SqlRepo {
	db, err := gorm.Open("sqlite3", name)
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Customer{},

		&models.Order{}).Error
	require.NoError(t, err)

	return &repo.SqlRepo{DB: db}
}
