package pubsub

type MessageCallback func(c PubSubClient, m []byte)
type ClientCallback func(c PubSubClient)

// PubSubClient interface
type PubSubClient interface {
	// connect to server
	Connect()

	// on connect callback
	OnConnect(callback ClientCallback)

	// publish event
	Publish(event string, mesage interface{})

	// Subscribe event
	Subscribe(event string, handler MessageCallback)

	// Unsubscribe event
	Unsubscribe(event ...string)

	// Check if client connected
	IsConnected() bool
}
