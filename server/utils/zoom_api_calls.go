package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func GetMeetingInfo(meetingId int64, authToken string) (*MeetingInfo, error) {

	tries := 0
	var err error

	for tries < 2 {
		time.Sleep(time.Second * 2)
		url := fmt.Sprintf("https://api.zoom.us/v2/meetings/%d", meetingId)

		req, er := http.NewRequest("GET", url, nil)
		if er != nil {
			tries += 1
			err = er
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		client := &http.Client{}
		resp, er := client.Do(req)
		if er != nil {
			tries += 1
			err = er
			continue
		}
		defer resp.Body.Close()
		body, er := ioutil.ReadAll(resp.Body)
		if er != nil {
			tries += 1
			err = er
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode > 300 {
			err = fmt.Errorf("received status code: %d from zoom, err: %s", resp.StatusCode, string(body))
			tries += 1
			continue
		}

		meetingInfo := &MeetingInfo{}

		if unmarshalErr := json.Unmarshal(body, meetingInfo); unmarshalErr != nil {
			log.Println("Error unmarshalling zoom meeting info from zoom in to struct ", unmarshalErr)
			// This error is not retryable so return success to webhook
			return nil, unmarshalErr
		}
		return meetingInfo, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("unexpected end of function")
}
