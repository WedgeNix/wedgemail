package wedgemail

import (
	"errors"
	"log"
	"strings"

	"github.com/OuttaLineNomad/throttle"

	gmail "google.golang.org/api/gmail/v1"
)

// Label adds labels to messages.
func (ms *MailService) Label(msgs []*gmail.Message, label []string) (err error) {
	list, err := ms.Service.Users.Labels.List("me").Do()
	if err != nil {
		return err
	}

	listMap := map[string]string{}
	for _, label := range list.Labels {
		lName := strings.ToLower(label.Name)
		listMap[lName] = label.Id
	}
	sendLabel := []string{}
	for _, l := range label {
		if id, ok := listMap[strings.ToLower(l)]; ok {
			sendLabel = append(sendLabel, id)
		} else {
			return errors.New(`the label "` + l + `" you provided; was not found in lables`)
		}
	}

	ok := &gmail.ModifyMessageRequest{
		AddLabelIds: sendLabel,
	}
	gmsg := &gmail.Message{}
	for _, msg := range msgs {
		err := backoff.Run(func() (err error) {
			gmsg, err = ms.Service.Users.Messages.Modify("me", msg.Id, ok).Do()
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
		// err := expDo(ms.Service.Users.Messages.Modify("me", msg.Id, ok).Do, gmsg)
		if err != nil {
			log.Panic(err)
		}
	}
	return nil
}
