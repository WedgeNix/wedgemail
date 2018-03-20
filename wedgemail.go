package wedgemail

import (
	"context"
	"encoding/base64"
	"errors"
	"math"
	"math/rand"
	"net/mail"
	"reflect"
	"strings"
	"time"

	"google.golang.org/api/googleapi"

	"io/ioutil"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Injection injects structure into email template.
type injection struct {
	Vendor string
	Date   string
	PO     string
}

// MailService holds gmail service passes to methods
type MailService struct {
	From    *mail.Address
	Service *gmail.Service
}

// StartMail starts wedgemail
func StartMail() (*MailService, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("credentials/client_secret.json")
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, gmail.MailGoogleComScope)
	if err != nil {
		return nil, err
	}
	client := getClient(ctx, config)

	srv, err := gmail.New(client)
	if err != nil {
		return nil, err
	}
	return &MailService{Service: srv}, nil

}

type wrapper struct {
	v interface{}
	error
	attempts float64
}

func (w *wrapper) wrap(i interface{}, err error) *wrapper {
	w.v = i
	w.error = err
	return w
}

// // expDo is an exponential backoff func to use with Do() for gmail api.
// func (w *wrapper) expDo(v interface{}, err *error) bool {
// 	if w.error != nil {
// 		five := strings.Contains(w.error.Error(), "500")
// 		four := strings.Contains(w.error.Error(), "429")
// 		if five || four {
// 			maxWait := 48000
// 			wait := int(math.Min(float64(maxWait), math.Pow(2, w.attempts)+float64(rand.Intn(1000))+1))
// 			time.Sleep(time.Duration(wait) * time.Millisecond)
// 			w.attempts++
// 			if wait == maxWait {
// 				*err = errors.New("Attempts hit max")
// 				return false
// 			}
// 			return true
// 		}
// 		*err = w.error
// 		return false
// 	}
// 	if v == nil {
// 		return false
// 	}
// 	switch ptr := v.(type) {
// 	case *gmail.ListMessagesResponse:
// 		*ptr = *w.v.(*gmail.ListMessagesResponse)
// 	case *gmail.Message:
// 		*ptr = *w.v.(*gmail.Message)
// 	default:
// 		panic("Unknown Type for expDo()")
// 	}
// 	return false
// }

// expDo is an exponential backoff func to use with Do() for gmail api.
func expDo(f interface{}, ret interface{}, options ...googleapi.CallOption) error {
	var attempts float64
	const maxWait = 48000
	var wait int
	vf := reflect.ValueOf(f)
	var ops []reflect.Value
	for _, op := range options {
		ops = append(ops, reflect.ValueOf(op))
	}
	for wait < maxWait {
		result := vf.Call(ops)
		ret = result[0].Interface()

		if err := result[1].Interface().(error); err != nil {
			five := strings.Contains(err.Error(), "500")
			four := strings.Contains(err.Error(), "429")
			if five || four {
				wait = int(math.Min(float64(maxWait), math.Pow(2, attempts)+float64(rand.Intn(1000))+1))
				time.Sleep(time.Duration(wait) * time.Millisecond)
				attempts++
				continue
			}
			return err
		}
		return nil
	}
	return errors.New("Attempts hit max")
}

func encodeWeb64String(b []byte) string {

	s := base64.URLEncoding.EncodeToString(b)

	var i = len(s) - 1
	for s[i] == '=' {
		i--
	}

	return s[0 : i+1]
}
