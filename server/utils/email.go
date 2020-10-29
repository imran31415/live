package utils

import (
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"log"
)

// TODO (refactor for reuse if needed, currently deprecated
func SendInstructorInvitation(recipientName string, recipientEmail string, owner string, ownerEmail string, providerName string, instructorInvitationUrl string) error {

	m := mail.NewV3Mail()

	address := "contact@via.live"
	name := "Via Live"
	e := mail.NewEmail(name, address)
	m.SetFrom(e)
	m.SetTemplateID("d-57bf013a8c0a497c826a58107f66f5f9")

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(recipientName, recipientEmail),
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("recipient_name", recipientName)
	p.SetDynamicTemplateData("recipient_email", recipientEmail)
	p.SetDynamicTemplateData("subject", fmt.Sprintf("You are invited to join the %s team on VIA.LIVE!", providerName))
	p.SetDynamicTemplateData("owner", owner)
	p.SetDynamicTemplateData("owner_email", ownerEmail)
	p.SetDynamicTemplateData("provider_name", providerName)
	p.SetDynamicTemplateData("instructor_invitation_url", instructorInvitationUrl)

	m.AddPersonalizations(p)

	request := sendgrid.GetRequest("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", "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode > 300 {
		return fmt.Errorf("received non 200 response code from sendgrid, code: %d, body: %s", response.StatusCode, response.Body)
	}
	log.Println("Successfully sent email.")
	return nil

}

//
//func SendPaymentConfirmation(recipient_name string, recipient_email string, pc_subject string, pc_paid_amount string, pc_session_name string, pc_recipient string, pc_zoom_url string) {
//	m := mail.NewV3Mail()
//
//	address := "contact@via.live"
//	name := "Via Live"
//	e := mail.NewEmail(name, address)
//	m.SetFrom(e)
//	m.SetTemplateID("d-66071e346b204d06b88e3e4c1b30bdde")
//
//	p := mail.NewPersonalization()
//	tos := []*mail.Email{
//		mail.NewEmail(recipient_name, recipient_email),
//	}
//	p.AddTos(tos...)
//
//	p.SetDynamicTemplateData("paid_amount", pc_paid_amount)
//	p.SetDynamicTemplateData("session_name", pc_session_name)
//	p.SetDynamicTemplateData("recipient", pc_recipient)
//	p.SetDynamicTemplateData("zoom_url", pc_zoom_url)
//	p.SetDynamicTemplateData("subject", pc_subject)
//
//	m.AddPersonalizations(p)
//
//	request := sendgrid.GetRequest("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", "/v3/mail/send", "https://api.sendgrid.com")
//	request.Method = "POST"
//	var Body = mail.GetRequestBody(m)
//	request.Body = Body
//	response, err := sendgrid.API(request)
//	if err != nil {
//		fmt.Println(err)
//	} else {
//		fmt.Println(response.StatusCode)
//		fmt.Println(response.Body)
//		fmt.Println(response.Headers)
//	}
//}
//
//func SendSessionScheduledConfirmation(recipient_name string, recipient_email string, session_date string, session_name string, event_url string, subject string) {
//	m := mail.NewV3Mail()
//
//	address := "contact@via.live"
//	name := "Via Live"
//	e := mail.NewEmail(name, address)
//	m.SetFrom(e)
//	m.SetTemplateID("d-ff9dd3a0f96e40f4ba62ed63dee4fa8f")
//
//	p := mail.NewPersonalization()
//	tos := []*mail.Email{
//		mail.NewEmail(recipient_name, recipient_email),
//	}
//	p.AddTos(tos...)
//
//	p.SetDynamicTemplateData("session_date", session_date)
//	p.SetDynamicTemplateData("session_name", session_name)
//	p.SetDynamicTemplateData("event_url", event_url)
//	p.SetDynamicTemplateData("subject", subject)
//
//	m.AddPersonalizations(p)
//
//	request := sendgrid.GetRequest("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", "/v3/mail/send", "https://api.sendgrid.com")
//	request.Method = "POST"
//	var Body = mail.GetRequestBody(m)
//	request.Body = Body
//	response, err := sendgrid.API(request)
//	if err != nil {
//		fmt.Println(err)
//	} else {
//		fmt.Println(response.StatusCode)
//		fmt.Println(response.Body)
//		fmt.Println(response.Headers)
//	}
//
//}
