package grpc

import (
	"admin/models"
	pb "admin/protos"
	"admin/server/repo"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"net"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	stripe "github.com/stripe/stripe-go/v71"
)

const (
	paymentMethodDefault = "card"
	signerUrl            = "https://livehub-277906.uc.r.appspot.com/sign"
	bucketName           = "image_distro"
)

// Database abstracts out database specific logic from GRPC server logic.
// If we want to switch from GORM, we just need a new repo that satisfies the below methods
type Database interface {
	GetUserById(id uint) (*models.User, error)
	GetUsersById(ids []uint) (models.Users, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByZoomAccountId(id string) (*models.User, error)
	GetOrderById(id uint) (*models.Order, error)
	GetOrderByPaymentPlatformOrderId(id string) (*models.Order, error)
	GetCustomerById(id uint) (*models.Customer, error)
	GetOrCreateUserBySubId(id string) (*models.User, error)
	UpdateUserBySubId(u *models.User) (*models.User, error)
	UpdateUserById(u *models.User) (*models.User, error)
	UpdateUserZoomDeAuthorized(id uint) error

	// Supercedes Classes
	GetSessionById(id uint) (*models.Session, error)
	GetSessionByUserIdAndZoomMeetingId(userId, zoomMeetingId int64) (*models.Session, error)
	GetSessionByUserIdAndZoomMeetingIdAndOccurrenceId(userId, zoomMeetingId int64, occurrenceId string) (*models.Session, error)
	GetSessionByUserIdAndZoomMeetingIdAndRawPayloadId(userId, zoomMeetingId int64, zoomRawPayload string) (*models.Session, error)
	DeleteSessionById(id uint) error
	DeleteSessionsByUserIdAndZoomMeetingId(userId, zoomMeetingId uint) error
	DeleteSessionsByUserIdAndZoomMeetingIdAndNotOccurrenceIds(userId, zoomMeetingId uint, notOccurrenceIds []string) error
	GetSessionsByStartDate(date int64, limit int) (models.Sessions, error)
	GetSessionsByStartDateAndTag(date int64, limit int, tag string) (models.Sessions, error)
	GetSessionsByUserId(id uint) (models.Sessions, error)
	GetSessionByUserIdAndNameAndDateAndDuration(userId, date, duration int64, name string) (*models.Session, error)
	GetSessionByUserIdAndMeetingId(userId, meetingId int64) (*models.Session, error)
	UpdateSession(c *models.Session) (*models.Session, error)
	IsSessionIsPurchasedByUser(userId, SessionId uint) (bool, error)
	CreateSession(Session *models.Session) (*models.Session, error)
	GetPreviousSessionsByUserIdAndDate(id uint, date int64) (models.Sessions, error)
	GetUpcomingSessionsByUserIdAndDate(id uint, date int64) (models.Sessions, error)
	GetPreviousSessionsByIds(ids []uint) (models.Sessions, error)
	GetUpcomingSessionsByIds(ids []uint) (models.Sessions, error)
	UpdateSessionsNoZoomSyncByUserId(userId int64) error

	GetOrCreateCustomerByUserId(id uint) (*models.Customer, error)
	GetCustomersByIds(ids []uint) (models.Customers, error)
	UpdateCustomer(c *models.Customer) (*models.Customer, error)
	GetSucceededOrdersByCustomerId(id uint) (models.Orders, error)
	CreateOrder(o *models.Order) (*models.Order, error)
	UpdateOrderStatusByOrderId(o *models.Order) (*models.Order, error)
	UpdateOrderPaymentMethodId(orderId uint, paymentMethodId string) (*models.Order, error)

	CreateImage(image *models.Image) (*models.Image, error)
	GetImageByObjectId(objectId string) (*models.Image, error)
	UpdateImageStatus(image *models.Image) (*models.Image, error)
	GetImagesByUserId(id uint) (models.Images, error)

	GetZoomTokenById(id uint) (*models.ZoomToken, error)
	GetZoomTokenByUserId(id uint) (*models.ZoomToken, error)
	UpdateZoomTokenById(token *models.ZoomToken) (*models.ZoomToken, error)
	CreateZoomAccessToken(z *models.ZoomToken) (*models.ZoomToken, error)
	DeleteZoomAccessTokensByUserId(userId int64) error

	Close() error
}

type Server struct {
	repo                   Database
	stripeKey              string
	zoomClientKey          string
	zoomClientSecret       string
	zoomRedirectUri        string
	zoomreDirectSuccessUri string
}

func NewServer(host, name, pass, user, zoomClientKey, zoomClientSecret, zoomRedirectUri, zoomreDirectSuccessUri string) (*Server, error) {
	r, err := repo.NewSqlRepo(host, name, pass, user)
	if err != nil {
		return nil, err
	}

	return NewServerWithRepo(r, zoomClientKey, zoomClientSecret, zoomRedirectUri, zoomreDirectSuccessUri), nil
}

func NewServerWithRepo(r *repo.SqlRepo, zoomClientKey, zoomClientSecret, zoomRedirectUri, zoomreDirectSuccessUri string) *Server {
	return &Server{
		repo:                   r,
		zoomClientKey:          zoomClientKey,
		zoomClientSecret:       zoomClientSecret,
		zoomRedirectUri:        zoomRedirectUri,
		zoomreDirectSuccessUri: zoomreDirectSuccessUri,
	}
}

func Run(grpcPort, host, name, pass, user, stripeKey, zoomClientKey, zoomClientSecret, zoomRedirectUri, zoomreDirectSuccessUri string) error {
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to listen: %v", err))
	}

	s := grpc.NewServer()
	serv, er := NewServer(host, name, pass, user, zoomClientKey, zoomClientSecret, zoomRedirectUri, zoomreDirectSuccessUri)
	// set stripe key with strip library
	stripe.Key = stripeKey
	if er != nil {
		return errors.New(fmt.Sprintf("err setting up db: %s", er))
	}

	// interface compliance of GRPC server to protobuf checked here:
	pb.RegisterNemoServer(s, serv)

	if err = s.Serve(lis); err != nil {
		return err
	}
	return nil
}
