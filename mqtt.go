package pubsub

import (
	"crypto/tls"
	"log"
	"net/url"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTConfig struct {
	URL                  string
	Debug                bool
	DisableAutoReconnect bool
	KeepAliveSecond      int64
	PingTimeoutSecond    int64
	ReconnectInterval    int64
}

type MQTTClient struct {
	client      *mqtt.Client
	subscribers map[string]MessageCallback
	onConnect   ConnectCallback
	config      *MQTTConfig
}

func (m *MQTTClient) Connect() {
	if nil == m.config || m.config.URL == "" {
		return
	}
	if m.config.KeepAliveSecond == 0 {
		m.config.KeepAliveSecond = 10
	}
	if m.config.PingTimeoutSecond == 0 {
		m.config.PingTimeoutSecond = 10
	}
	if m.config.ReconnectInterval == 0 {
		m.config.ReconnectInterval = 3
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.config.URL)
	opts.SetAutoReconnect(!m.config.DisableAutoReconnect)
	opts.SetKeepAlive(time.Duration(m.config.KeepAliveSecond) * time.Second)
	opts.SetPingTimeout(time.Duration(m.config.KeepAliveSecond) * time.Second)
	if !m.config.DisableAutoReconnect {
		opts.SetMaxReconnectInterval(time.Duration(m.config.ReconnectInterval) * time.Second)
	}
	opts.SetConnectionLostHandler(func(c mqtt.Client, e error) {
		m.uninitialize()
		if m.config.Debug {
			log.Println("[MQTT] Disconnected")
		}
	})
	opts.SetReconnectingHandler(func(c mqtt.Client, co *mqtt.ClientOptions) {
		if m.config.Debug {
			log.Println("[MQTT] Reconnecting")
		}
	})
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		m.initialize(&c)
		if m.config.Debug {
			log.Println("[MQTT] Connected")
		}
	})

	opts.SetConnectRetry(!m.config.DisableAutoReconnect)
	if !m.config.DisableAutoReconnect {
		opts.SetConnectRetryInterval(time.Duration(m.config.ReconnectInterval) * time.Second)
	}
	opts.SetConnectionAttemptHandler(func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
		if m.config.Debug {
			log.Println("[MQTT] Connecting")
		}
		return tlsCfg
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		if m.config.Debug {
			log.Println("[MQTT] Disconnected")
		}
	}
}

func (m *MQTTClient) Subscribe(topic string, handler MessageCallback) {
	if nil == m.subscribers {
		m.subscribers = map[string]MessageCallback{}
	}
	m.subscribers[topic] = handler
}

func (m MQTTClient) Unsubscribe(topics ...string) {
	if nil == m.client || !(*m.client).IsConnected() {
		return
	}

	cl := *m.client
	cl.Unsubscribe(topics...)
	for t := range topics {
		var topic string = topics[t]
		delete(m.subscribers, topic)
	}
}

func (m MQTTClient) Publish(topic string, message interface{}) {
	if nil == m.client || !(*m.client).IsConnected() {
		return
	}

	cl := *m.client
	cl.Publish(topic, 0, false, message)
}

func (m *MQTTClient) initialize(c *mqtt.Client) {
	if nil != c {
		m.client = c
	}

	if nil == m.client || !(*m.client).IsConnected() {
		return
	}

	cl := *m.client
	var mc Client = m

	if nil != m.onConnect {
		go func() {
			m.onConnect(mc)
		}()
	}

	for topic, handler := range m.subscribers {
		token := cl.Subscribe(topic, 0, func(c mqtt.Client, r mqtt.Message) {
			handler(mc, r.Payload())
		})
		if token.Wait() && token.Error() != nil {
			log.Println(token.Error().Error())
		}
	}
}

func (m *MQTTClient) OnConnect(callback ConnectCallback) {
	m.onConnect = callback
}

func (m *MQTTClient) uninitialize() {
	m.client = nil
}

func (m MQTTClient) IsConnected() bool {
	return nil != m.client && (*m.client).IsConnected()
}

func NewMQTTClient(config *MQTTConfig) Client {
	var mc Client = &MQTTClient{
		config: config,
	}
	return mc
}
