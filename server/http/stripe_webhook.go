package http

import (
	pb "admin/protos"
	"encoding/json"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/webhook"
	"io/ioutil"
	"log"
	"net/http"
)

var StripeWebHookHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, wErr := webhook.ConstructEvent(payload, req.Header.Get(stripesignatureHeader),
		stripeWebHookSecret)

	if wErr != nil {
		log.Printf("Error verifying webhook signature: %v\n", wErr)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err = json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Println("Order succeeded payment intent id: ", paymentIntent.ID)
		updateOrderStatus(paymentIntent.ID, pb.Order_SUCCEEDED, w)
		updateOrderPaymentMethodId(paymentIntent.ID, paymentIntent.PaymentMethod.ID, w)

		return

	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		err = json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Println("PaymentIntent failed!, intent Id: ", paymentIntent.ID)
		updateOrderStatus(paymentIntent.ID, pb.Order_FAILED, w)
	case "payment_intent.created":
		var paymentIntent stripe.PaymentIntent
		err = json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Println("PaymentIntent Created!, intent Id: ", paymentIntent.ID)
		updateOrderStatus(paymentIntent.ID, pb.Order_PROVIDER_CREATED, w)

	default:
		log.Printf("Unexpected event type, skipping: %s\n", event.Type)
		w.WriteHeader(http.StatusOK)
		return
	}
	return
})
