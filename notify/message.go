package notify

import (
	"firebase.google.com/go/messaging"
)

func createMessage(topic, title, body string, data map[string]string) *messaging.Message {
	return &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}
}

func createMessageWithCondition(condition, title, body string, data map[string]string) *messaging.Message {
	return &messaging.Message{
		Condition: condition,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}
}
