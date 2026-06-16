// Package entity defines WebSocket message contracts for the AlphaStream streaming API.
// All messages are JSON-serializable and typed via WsMessageType constants.
package entity

import "time"

// ─── WebSocket Message Types ──────────────────────────────────────────────────

// WsMessageType is a string enum for all WebSocket message types.
// Using typed string constants prevents accidental misuse of raw strings.
type WsMessageType string

const (
	// WsMsgOHLCVUpdate is sent when a new candlestick closes or is updated.
	WsMsgOHLCVUpdate WsMessageType = "OHLCV_UPDATE"

	// WsMsgTickerUpdate is sent on every price tick (sub-candle update).
	WsMsgTickerUpdate WsMessageType = "TICKER_UPDATE"

	// WsMsgIndicatorUpdate is sent after new indicator values are computed.
	WsMsgIndicatorUpdate WsMessageType = "INDICATOR_UPDATE"

	// WsMsgPredictionUpdate is sent when the prediction engine produces a new result.
	WsMsgPredictionUpdate WsMessageType = "PREDICTION_UPDATE"

	// WsMsgSubscribe is sent by the client to subscribe to a symbol feed.
	WsMsgSubscribe WsMessageType = "SUBSCRIBE"

	// WsMsgUnsubscribe is sent by the client to unsubscribe from a symbol feed.
	WsMsgUnsubscribe WsMessageType = "UNSUBSCRIBE"

	// WsMsgError is sent by the server to signal a processing error to the client.
	WsMsgError WsMessageType = "ERROR"

	// WsMsgPing / WsMsgPong are used for connection health checks.
	WsMsgPing WsMessageType = "PING"
	WsMsgPong WsMessageType = "PONG"
)

// ─── WebSocket Envelope ───────────────────────────────────────────────────────

// WsMessage is the top-level envelope for all WebSocket messages.
// The Payload field is typed as interface{} so it can carry any domain struct.
// Clients should switch on the Type field before deserializing Payload.
type WsMessage struct {
	Type      WsMessageType `json:"type"`
	Symbol    string        `json:"symbol,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Payload   interface{}   `json:"payload"`
}

// WsErrorPayload is the Payload for WsMsgError messages.
type WsErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WsSubscribePayload is the Payload sent by a client for SUBSCRIBE/UNSUBSCRIBE.
type WsSubscribePayload struct {
	Symbol    string `json:"symbol"`
	Timeframe string `json:"timeframe"` // e.g. "1m", "5m"
}

// ─── Broadcast Helper ─────────────────────────────────────────────────────────

// NewWsMessage constructs a typed WsMessage with the current UTC timestamp.
// This is a convenience constructor to ensure consistent message formatting.
func NewWsMessage(msgType WsMessageType, symbol string, payload interface{}) WsMessage {
	return WsMessage{
		Type:      msgType,
		Symbol:    symbol,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
}

// NewWsErrorMessage constructs a WsMsgError envelope.
func NewWsErrorMessage(code, message string) WsMessage {
	return NewWsMessage(WsMsgError, "", WsErrorPayload{
		Code:    code,
		Message: message,
	})
}
