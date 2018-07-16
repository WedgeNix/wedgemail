package wedgemail

import (
	"encoding/base64"
	"strings"

	"github.com/OuttaLineNomad/throttle"
	"google.golang.org/api/gmail/v1"
)

// AttachemtFiles holds filename and data from emails for user to deal with
type AttachemtFiles struct {
	Filename string
	Data     []byte
}

// GetQuery gets a list of messages matching query.
func (ms *MailService) GetQuery(q string) (*gmail.ListMessagesResponse, error) {
	// msgs := &gmail.ListMessagesResponse{}
	msgs, err := ms.Service.Users.Messages.List("me").Q(q).Do()
	return msgs, err
}

// GetAttachments gets slice of AttachemtFiles which has name and bites for each file
func (ms *MailService) GetAttachments(list []*gmail.Message, ext []string) ([]AttachemtFiles, error) {
	attchs := []AttachemtFiles{}
	for _, msg := range list {
		msgID := msg.Id

		var usrMsg *gmail.Message
		err := backoff.Run(func() (err error) {
			usrMsg, err = ms.Service.Users.Messages.Get("me", msgID).Do()
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
		// err := expDo(ms.Service.Users.Messages.Get("me", msgID).Do, &usrMsg)
		if err != nil {
			return nil, err
		}

		for _, part := range usrMsg.Payload.Parts {
			attID := part.Body.AttachmentId
			if attID == "" {
				continue
			}
			attName := part.Filename
			if !findExt(attName, ext) {
				continue
			}
			var msgBody *gmail.MessagePartBody
			err := backoff.Run(func() (err error) {
				msgBody, err = ms.Service.Users.Messages.Attachments.Get("me", msgID, attID).Do()
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
			// err := expDo(ms.Service.Users.Messages.Attachments.Get("me", msgID, attID).Do, &msgBody)
			if err != nil {
				return nil, err
			}
			b, err := base64.URLEncoding.DecodeString(msgBody.Data)
			if err != nil {
				return nil, err
			}
			attFile := AttachemtFiles{
				Filename: attName,
				Data:     b,
			}
			attchs = append(attchs, attFile)
		}
	}
	return attchs, nil
}

func findExt(name string, exts []string) bool {
	want := strings.Split(name, ".")
	for _, ext := range exts {
		if want[len(want)-1] == strings.ToLower(ext) {
			return true
		}
	}
	return false
}
