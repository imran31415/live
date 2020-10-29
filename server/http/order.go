package http

import (
	pb "admin/protos"
	"context"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"io/ioutil"
	"log"
	"net/http"
)

var CreateOrder = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}

	u, err := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: authZeroSubId})
	if err != nil {
		log.Println("Error getting user: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	in := &pb.CreateOrderRequest{}
	body, rErr := ioutil.ReadAll(r.Body)
	if rErr != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, rErr.Error(), http.StatusBadRequest)
		return
	}
	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	in.UserId = u.GetId()

	out, cErr := httpClient.grpcClient.CreateOrder(context.Background(), in)
	if cErr != nil {
		log.Println("Error creating order: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}
	js, mErr := json.Marshal(out)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

func updateOrderStatus(paymentIntentId string, status pb.Order_PaymentPlatformStatus, w http.ResponseWriter) {
	o, oErr := httpClient.grpcClient.GetOrderByPaymentPlatformOrderId(context.Background(), &pb.PaymentPlatformOrderId{Id: paymentIntentId})
	if oErr != nil {
		log.Printf("Error getting payment_intent_id: %s in GetOrderByPaymentPlatformOrderId: %v\n", paymentIntentId, oErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, uErr := httpClient.grpcClient.UpdateOrderStatusByOrderId(context.Background(), &pb.Order{Id: o.GetId(), Status: status}); uErr != nil {
		log.Printf("Error in UpdateOrderStatusByOrderId: %v\n", uErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func updateOrderPaymentMethodId(paymentIntentId string, paymentMethodId string, w http.ResponseWriter) {
	o, oErr := httpClient.grpcClient.GetOrderByPaymentPlatformOrderId(context.Background(), &pb.PaymentPlatformOrderId{Id: paymentIntentId})
	if oErr != nil {
		log.Printf("Error getting payment_intent_id: %s in GetOrderByPaymentPlatformOrderId: %v\n", paymentIntentId, oErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, uErr := httpClient.grpcClient.UpdateOrderPaymentMethodId(context.Background(), &pb.UpdateOrderPaymentMethodIdRequest{
		OrderId:         o.GetId(),
		PaymentMethodId: paymentMethodId,
	}); uErr != nil {
		log.Printf("Error in UpdateOrderStatusByOrderId: %v\n", uErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
