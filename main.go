package main

import (
	"admin/server/grpc"
	"admin/server/http"
	"log"
	"os"
)

const (
	grpcPort = ":50051"
)

func main() {

	var host, name, pass, user, stripeKey, stripeWebhookSecret, zoomClientKey, zoomClientSecret, zoomRedirectUri, zoomreDirectSuccessUri string
	host = getEnvVar("DB_HOST", "localhost")
	name = getEnvVar("DB_NAME", "go-admin-test")
	pass = getEnvVar("DB_PASS", "ttianjun")
	user = getEnvVar("DB_USER", "root")
	stripeKey = getEnvVar("STRIPE_KEY", "")
	stripeWebhookSecret = getEnvVar("STRIPE_WEBHOOK_SECRET", "")
	zoomClientKey = getEnvVar("ZOOM_CLIENT_KEY", "")
	zoomClientSecret = getEnvVar("ZOOM_CLIENT_SECRET", "")
	zoomRedirectUri = getEnvVar("ZOOM_REDIRECT_URI", "")
	zoomreDirectSuccessUri = getEnvVar("ZOOM_REDIRECT_SUCCESS_URI", "")

	// Run the GRPC server
	//  note: use go routine so it doesn't block
	go func() {
		log.Println("Running GRPC Server....")
		if err := grpc.Run(grpcPort, host, name, pass, user, stripeKey, zoomClientKey, zoomClientSecret, zoomRedirectUri, zoomreDirectSuccessUri); err != nil {
			log.Printf("Error running grpc server, err: %s", err)
		}
	}()

	log.Println("Running HTTP Server...")
	err := http.Run(stripeWebhookSecret, zoomClientKey, zoomClientSecret)
	if err != nil {
		log.Println("Error running Http server: ", err)
	}

}

func getEnvVar(key, defaultValue string) string {
	i, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return i
}
