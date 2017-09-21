package wedgemail

import (
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/mail"
	"strings"
	"time"

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
	Service *gmail.Service
}

// StartMail starts wedgemail
func StartMail() (*MailService, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("credentials/client_secret.json")
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope, gmail.GmailLabelsScope)
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

// Email sends email with attachments.
func (ms *MailService) Email(to []string, subject string, content string, fileName ...string) error {
	from := mail.Address{Name: "WedgeNix", Address: "wedgenix.customercare@gmail.com"}
	toStr := strings.Join(to, ",")
	var message gmail.Message

	boundary := "__WedgeNix_Server_Mailing__"

	var attachments string
	for _, name := range fileName {
		if name == "" {
			continue
		}
		fileBytes, err := ioutil.ReadFile(name)
		fileMIMEType := http.DetectContentType(fileBytes)
		fileData := base64.StdEncoding.EncodeToString(fileBytes)
		if err != nil {
			return err
		}
		attachments += "--" + boundary + "\n" +
			"Content-Type: " + fileMIMEType + "; name=" + string('"') + name + string('"') + " \n" +
			"MIME-Version: 1.0\n" +
			"Content-Transfer-Encoding: base64\n" +
			"Content-Disposition: attachment; filename=" + string('"') + name + string('"') + " \n\n" +
			chunkSplit(fileData, 76, "\n")
	}

	messageBody := []byte("Content-Type: multipart/mixed; boundary=" + boundary + " \n" +
		"MIME-Version: 1.0\n" +
		"to: " + toStr + "\n" +
		"from: " + from.String() + "\n" +
		"subject: " + subject + "\n\n" +

		"--" + boundary + "\n" +
		"Content-Type: text/html; charset=" + string('"') + "UTF-8" + string('"') + "\n" +
		"MIME-Version: 1.0\n" +
		"Content-Transfer-Encoding: 7bit\n\n" +
		content + "\n\n" +

		attachments +

		"--" + boundary + "--")

	message.Raw = base64.URLEncoding.EncodeToString(messageBody)
	attempts := 0.0
	for {
		_, err := ms.Service.Users.Messages.Send("me", &message).Do()
		if err != nil {
			five := strings.Contains(err.Error(), "500")
			four := strings.Contains(err.Error(), "429")
			if five || four {
				maxWait := 48000
				wait := int(math.Min(float64(maxWait), math.Pow(2, attempts)+float64(rand.Intn(1000))+1))
				time.Sleep(time.Duration(wait) * time.Millisecond)
				attempts++
				if wait == maxWait {
					return errors.New("Attempts hit max")
				}
				continue
			}
			return err
		}
		return nil
	}
}

func encodeWeb64String(b []byte) string {

	s := base64.URLEncoding.EncodeToString(b)

	var i = len(s) - 1
	for s[i] == '=' {
		i--
	}

	return s[0 : i+1]
}

func chunkSplit(body string, limit int, end string) (res string) {

	var charSlice []rune

	// push characters to slice
	for _, char := range body {
		charSlice = append(charSlice, char)
	}

	defer func() {
		recover()
		res = string(charSlice) + end
	}()

	for len(charSlice) >= 1 {
		// convert slice/array back to string
		// but insert end at specified limit

		res = res + string(charSlice[:limit]) + end

		// discard the elements that were copied over to result
		charSlice = charSlice[limit:]

		// change the limit
		// to cater for the last few words in
		//
		if len(charSlice) < limit {
			limit = len(charSlice)
		}

	}

	return

}
