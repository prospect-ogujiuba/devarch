package events

import "encoding/json"

func EncodePayload(payload any) (json.RawMessage, error) {
	if payload == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(encoded), nil
}
