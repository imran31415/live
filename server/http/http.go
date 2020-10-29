package http

import (
	pb "admin/protos"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

const (
	port        = ":50052"
	grpcAddress = "localhost:50051"

	stripesignatureHeader = "Stripe-Signature"
	grantTypeAuthorize    = "authorization_code"
	grantAuthorize        = "code"
	grantTypeRefresh      = "refresh_token"
	grantRefresh          = "refresh_token"
)

type HttpClient struct {
	grpcClient pb.NemoClient
}

var (
	httpClient          HttpClient
	stripeWebHookSecret string
	zoomClientKey       string
	zoomClientSecret    string
)

func Run(secret, zKey, zSecret string) error {
	stripeWebHookSecret = secret
	zoomClientKey = zKey
	zoomClientSecret = zSecret

	// Create a HTTP grpcClient against the GRPC server
	// TODO figure out the anon var here
	conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	c := pb.NewNemoClient(conn)

	httpClient = HttpClient{grpcClient: c}

	r := mux.NewRouter()
	r = addRoutes(r)
	handler := cors.AllowAll().Handler(r)
	return http.ListenAndServe(port, handler)
}

func addRoutes(r *mux.Router) *mux.Router {
	// Register Handlers for the HTTP server
	// Authed read endpoints
	//  - Note the user is inferred from the auth0 token
	r.Handle("/", Health)

	r.Handle("/get_user_profile", jwtMiddleWare().Handler(GetUserProfile)).Methods("GET", "OPTIONS")
	r.Handle("/get_user", jwtMiddleWare().Handler(GetUser)).Methods("GET", "OPTIONS")

	r.Handle("/get_user_by_id", jwtMiddleWare().Handler(GetUserByIdHandler)).Methods("POST", "OPTIONS")
	r.Handle("/get_user_by_id_noauth", GetUserByIdPublicHandler).Methods("POST", "OPTIONS")
	r.Handle("/update_user_by_sub_id", jwtMiddleWare().Handler(UpdateUserBySubId)).Methods("POST", "OPTIONS")

	r.Handle("/create_image", jwtMiddleWare().Handler(CreateImage)).Methods("POST", "OPTIONS")
	r.Handle("/get_signed_image_upload_url", jwtMiddleWare().Handler(GetSignedImageUploadUrl)).Methods("POST", "OPTIONS")
	r.Handle("/update_image_status", UpdateImageStatus).Methods("POST", "OPTIONS")
	r.Handle("/get_images_by_user_id", jwtMiddleWare().Handler(GetImagesByUserId)).Methods("GET", "OPTIONS")

	r.Handle("/create_order", jwtMiddleWare().Handler(CreateOrder)).Methods("POST", "OPTIONS")
	r.Handle("/stripe_webhook", StripeWebHookHandler).Methods("POST", "OPTIONS")

	r.Handle("/get_zoom_app_install_url", jwtMiddleWare().Handler(GetZoomAppInstallUrl)).Methods("POST", "OPTIONS")
	r.Handle("/zoom_app_install_url", ZoomAppInstallHandler).Methods("GET", "OPTIONS")
	r.Handle("/zoom_meeting_webhook", ZoomMeetingWebHook).Methods("POST", "OPTIONS")
	r.Handle("/zoom_deauth_webhook", ZoomDeAuthorizeWebHook).Methods("POST", "OPTIONS")
	r.Handle("/get_zoom_token_by_user", jwtMiddleWare().Handler(GetZoomTokenByUser)).Methods("GET", "OPTIONS")

	r.Handle("/get_stripe_app_install_url", jwtMiddleWare().Handler(GetStripeInstallUrl)).Methods("POST", "OPTIONS")

	r.Handle("/get_session_by_id_noauth", GetSessionByIdNoAuth).Methods("POST", "OPTIONS")
	r.Handle("/get_session_by_id", jwtMiddleWare().Handler(GetSessionByIdHandler)).Methods("POST", "OPTIONS")
	r.Handle("/delete_session_by_id", jwtMiddleWare().Handler(DeleteSessionById)).Methods("DELETE", "OPTIONS")
	r.Handle("/update_session", jwtMiddleWare().Handler(UpdateSession)).Methods("POST", "OPTIONS")
	r.Handle("/create_session", jwtMiddleWare().Handler(CreateSession)).Methods("POST", "OPTIONS")
	r.Handle("/get_sessions", jwtMiddleWare().Handler(GetSessionsByDate)).Methods("POST", "OPTIONS")

	r.Handle("/get_upcoming_sessions_by_user_and_date", jwtMiddleWare().Handler(GetUpcomingSessionsByUserIdAndDate)).Methods("POST", "OPTIONS")
	r.Handle("/get_previous_sessions_by_user_and_date", jwtMiddleWare().Handler(GetPreviousSessionsByUserIdAndDate)).Methods("POST", "OPTIONS")

	r.Handle("/get_sessions_noauth", GetSessionsByDate).Methods("POST", "OPTIONS")
	r.Handle("/get_sessions_by_tag", GetSessionsByDateAndTag).Methods("POST", "OPTIONS")
	r.Handle("/create_meeting_in_zoom", jwtMiddleWare().Handler(CreateMeetingInZoom)).Methods("POST", "OPTIONS")

	r.Handle("/verify_zoom", VerifyZoom).Methods("GET", "OPTIONS")
	r.Handle("/verifyZoom.html", VerifyZoom).Methods("GET", "OPTIONS")

	return r

}

var Health = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ok := struct {
		code int
		name string
	}{code: 200, name: "ok"}
	js, err := json.Marshal(ok)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)

})

func writeJs(js []byte, w http.ResponseWriter, r *http.Request) {
	log.Println("Successfully processed http request")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		return
	}
	w.Write(js)
	return
}
