package wedgemail

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/mail"
	"strings"

	"github.com/OuttaLineNomad/throttle"
	"google.golang.org/api/gmail/v1"
)

// Attachment to hold attachments for emails.
type Attachment struct {
	Name string
	io.Reader
}

// Email sends email with attachments.
func (ms *MailService) Email(to []string, subject string, content string, atts ...Attachment) (err error) {
	if ms.From == nil {
		ms.From = &mail.Address{Name: "WedgeNix", Address: "wedgenix.customercare@gmail.com"}
	}
	toStr := strings.Join(to, ",")
	var message gmail.Message

	boundary := "__WedgeNix_Server_Mailing__"

	var attachments string
	for _, att := range atts {
		if len(att.Name) == 0 || att.Reader == nil {
			continue
		}
		fileBytes, err := ioutil.ReadAll(att)
		fileMIMEType := http.DetectContentType(fileBytes)
		fileData := base64.StdEncoding.EncodeToString(fileBytes)
		if err != nil {
			return err
		}
		attachments += "--" + boundary + "\n" +
			"Content-Type: " + fileMIMEType + "; name=" + string('"') + att.Name + string('"') + " \n" +
			"MIME-Version: 1.0\n" +
			"Content-Transfer-Encoding: base64\n" +
			"Content-Disposition: attachment; filename=" + string('"') + att.Name + string('"') + " \n\n" +
			fileData + "\n"
	}

	messageBody := []byte("Content-Type: multipart/mixed; boundary=" + boundary + " \n" +
		"MIME-Version: 1.0\n" +
		"to: " + toStr + "\n" +
		"from: " + ms.From.String() + "\n" +
		"subject: " + subject + "\n\n" +

		"--" + boundary + "\n" +
		"Content-Type: text/html; charset=" + string('"') + "UTF-8" + string('"') + "\n" +
		"MIME-Version: 1.0\n" +
		"Content-Transfer-Encoding: 7bit\n\n" +
		content + "\n\n" +

		attachments +

		"--" + boundary + "--")
	message.Raw = base64.URLEncoding.EncodeToString(messageBody)
	err = backoff.Run(func() (err error) {
		_, err = ms.Service.Users.Messages.Send("me", &message).Do()
		if err != nil {
			five := strings.Contains(err.Error(), "500")
			four := strings.Contains(err.Error(), "429")
			if five || four {
				return err
			}
			throttle.NoGos(err)
		}
		return nil
	})
	// err = expDo(ms.Service.Users.Messages.Send("me", &message).Do, nil)
	return
}
