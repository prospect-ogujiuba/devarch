package events

import "encoding/json"

func MarshalEnvelope(envelope Envelope) ([]byte, error) {
	return json.Marshal(envelope)
}

func UnmarshalEnvelope(data []byte) (Envelope, error) {
	var envelope Envelope
	err := json.Unmarshal(data, &envelope)
	return envelope, err
}

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

func DecodePayload[T any](envelope Envelope) (T, error) {
	var payload T
	if len(envelope.Payload) == 0 {
		return payload, nil
	}
	err := json.Unmarshal(envelope.Payload, &payload)
	return payload, err
}
