package pubsub

type MessageCallback func(c Client, m []byte)
type ConnectCallback func(c Client)

// Client interface
type Client interface {
	// connect to server
	Connect()

	// on connect callback
	OnConnect(callback ConnectCallback)

	// publish event
	Publish(event string, mesage interface{})

	// Subscribe event
	Subscribe(event string, handler MessageCallback)

	// Unsubscribe event
	Unsubscribe(event ...string)

	// Check if client connected
	IsConnected() bool
}
