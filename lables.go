package wedgemail

import (
	"errors"
	"fmt"
	"log"

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
		listMap[label.Name] = label.Id
	}
	sendLabel := []string{}
	for _, l := range label {
		if id, ok := listMap[l]; ok {
			sendLabel = append(sendLabel, id)
		} else {
			return errors.New(`the lable "` + l + `" you provided; was not found in lables`)
		}
	}

	ok := &gmail.ModifyMessageRequest{
		AddLabelIds: sendLabel,
	}
	gmsg := &gmail.Message{}
	for _, msg := range msgs {
		err := expDo(ms.Service.Users.Messages.Modify("me", msg.Id, ok).Do, gmsg)
		if err != nil {
			log.Panic(err)
		}
	}
	fmt.Println(gmsg.Id)
	return nil
}
