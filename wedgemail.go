package wedgemail

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Injection injects structure into email template.
type injection struct {
	Vendor string
	Date   string
	PO     string
}

type MailService struct {
	Service *gmail.Service
}

// StartMail starts wedgemail
func StartMail() (*MailService, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("credentials/client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope, gmail.GmailLabelsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}
	return &MailService{Service: srv}, nil

}

// Email sends email with attachments.
func (ms *MailService) Email(to []string, subject string, content string, fileName ...string) error {
	from := mail.Address{Name: "WedgeNix", Address: "wedgenix.customercare@gmail.com"}
	toStr := mail.Address{Name: "", Address: strings.Join(to, ",")}
	var message gmail.Message

	boundary := "__WedgeNix_Server_Mailing__"

	var attachments string
	for _, name := range fileName {
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
		"to: " + toStr.String() + "\n" +
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

	_, err := ms.Service.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return err
	}
	return nil
}

func encodeWeb64String(b []byte) string {

	s := base64.URLEncoding.EncodeToString(b)

	var i = len(s) - 1
	for s[i] == '=' {
		i--
	}

	return s[0 : i+1]
}

func chunkSplit(body string, limit int, end string) string {

	var charSlice []rune

	// push characters to slice
	for _, char := range body {
		charSlice = append(charSlice, char)
	}

	result := ""

	for len(charSlice) >= 1 {
		// convert slice/array back to string
		// but insert end at specified limit

		result = result + string(charSlice[:limit]) + end

		// discard the elements that were copied over to result
		charSlice = charSlice[limit:]

		// change the limit
		// to cater for the last few words in
		//
		if len(charSlice) < limit {
			limit = len(charSlice)
		}

	}

	return result

}
