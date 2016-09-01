// Package voice is a demonstration of a home-grown speech recognition engine
// using a few Google technologies: Speech API, App Engine, Go.
package voice

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	// GAE has an old version of Go, hence the old version of net/context.
	"golang.org/x/net/context"
)

// In a non-GAE app, this would be the main func
func init() {
	http.HandleFunc("/", handler)
}

// These TwiML constants define the speech synthesis that users hear.
const (
	greeting = `<?xml version="1.0" encoding="UTF-8"?>
	<Response>
		<Say>Hello, human. What is the password?</Say>
		<Record timeout="5" />
	</Response>`
	badPassword = `<?xml version="1.0" encoding="UTF-8"?>
	<Response>
		<Say>No no no! That is not the password. I will now report you to Santa Claus.</Say>
	</Response>`
	okayPassword = `<?xml version="1.0" encoding="UTF-8"?>
	<Response>
		<Say>Yes, testing 1 2 3 is the password. I would now execute an arbitrary function or functions.</Say>
	</Response>`
)

// handler is the main routing endpoint
func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")

	// If there is no message, play the greeting.
	rec := r.FormValue("RecordingUrl")
	if rec == "" {
		fmt.Fprint(w, greeting)
		return
	}

	// Transcribe the audio and check the results
	c := appengine.NewContext(r)

	// Some additional logging
	for k, v := range r.PostForm {
		log.Infof(c, "%s: %v", k, v)
	}

	text, err := transcribe(c, rec)
	if err != nil {
		http.Error(w, "could not transcribe", http.StatusInternalServerError)
		log.Errorf(c, "could not transcribe: %v", err)
		return
	}
	if text == "testing 1 2 3" {
		fmt.Fprint(w, okayPassword)
	} else {
		fmt.Fprint(w, badPassword)
	}
}

// transcribe uses the audio from Twilio and checks against
// the Google Speech API transcription, using two helper functions:
// fetchAudio and fetchTranscription
func transcribe(c context.Context, url string) (string, error) {
	b, err := fetchAudio(c, url)
	if err != nil {
		return "", err
	}
	return fetchTranscription(c, b)
}

// fetchAudio retrieves the recorded audio from the Twilio API
func fetchAudio(c context.Context, url string) ([]byte, error) {
	client := urlfetch.Client(c)
	res, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch %v: %v", url, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetched with status: %s", res.Status)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response: %v", err)
	}
	return b, nil
}

var speechURL = "https://speech.googleapis.com/v1beta1/speech:syncrecognize?key=" +
	os.Getenv("SPEECH_API_KEY")

type speechRequest struct {
	Config struct {
		Encoding   string `json:"encoding"`
		SampleRate int    `json:"sampleRate"`
	} `json:"config"`
	Audio struct {
		Content string `json:"content"`
	} `json:"audio"`
}

// fetchTranscription sends the Twilio audio data as a base64-encoded string
// to the Google Speech API service and returns the requested transcription
func fetchTranscription(c context.Context, b []byte) (string, error) {
	var req speechRequest
	req.Config.Encoding = "LINEAR16"
	req.Config.SampleRate = 8000
	req.Audio.Content = base64.StdEncoding.EncodeToString(b)

	// Convert our speechRequest struct to JSON before sending to the Speech API
	jBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("could not encode speech request properly: %v", err)
	}
	res, err := urlfetch.Client(c).Post(speechURL, "application/json", bytes.NewReader(jBytes))
	if err != nil {
		return "", fmt.Errorf("could not transcribe audio: %v", err)
	}

	// Set the data types we need to decode from the Speech API response.
	var data struct {
		Error struct {
			Code    int
			Message string
			Status  string
		}
		Results []struct {
			Alternatives []struct {
				Transcript string
				Confidence float64
			}
		}
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("could not decode speech response: %v", err)
	}
	if data.Error.Code != 0 {
		return "", fmt.Errorf("Speech API error has occurred! %d %s %s",
			data.Error.Code, data.Error.Status, data.Error.Message)
	}
	if len(data.Results) == 0 || len(data.Results[0].Alternatives) == 0 {
		return "", fmt.Errorf("No transcriptions found.")
	}
	return data.Results[0].Alternatives[0].Transcript, nil
}
