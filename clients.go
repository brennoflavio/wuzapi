package main

import (
	"sync"

	"github.com/go-resty/resty/v2"
	"go.mau.fi/whatsmeow"
)

type ClientManager struct {
	sync.RWMutex
	whatsmeowClient *whatsmeow.Client
	httpClient      *resty.Client
	myClient        *MyClient
}

func NewClientManager() *ClientManager {
	return &ClientManager{}
}

func (cm *ClientManager) SetWhatsmeowClient(client *whatsmeow.Client) {
	cm.Lock()
	defer cm.Unlock()
	cm.whatsmeowClient = client
}

func (cm *ClientManager) GetWhatsmeowClient() *whatsmeow.Client {
	cm.RLock()
	defer cm.RUnlock()
	return cm.whatsmeowClient
}

func (cm *ClientManager) DeleteWhatsmeowClient() {
	cm.Lock()
	defer cm.Unlock()
	cm.whatsmeowClient = nil
}

func (cm *ClientManager) SetHTTPClient(client *resty.Client) {
	cm.Lock()
	defer cm.Unlock()
	cm.httpClient = client
}

func (cm *ClientManager) GetHTTPClient() *resty.Client {
	cm.RLock()
	defer cm.RUnlock()
	return cm.httpClient
}

func (cm *ClientManager) DeleteHTTPClient() {
	cm.Lock()
	defer cm.Unlock()
	cm.httpClient = nil
}

func (cm *ClientManager) SetMyClient(client *MyClient) {
	cm.Lock()
	defer cm.Unlock()
	cm.myClient = client
}

func (cm *ClientManager) GetMyClient() *MyClient {
	cm.RLock()
	defer cm.RUnlock()
	return cm.myClient
}

func (cm *ClientManager) DeleteMyClient() {
	cm.Lock()
	defer cm.Unlock()
	cm.myClient = nil
}

// UpdateMyClientSubscriptions updates the event subscriptions of a client without reconnecting
func (cm *ClientManager) UpdateMyClientSubscriptions(subscriptions []string) {
	cm.Lock()
	defer cm.Unlock()
	if cm.myClient != nil {
		cm.myClient.subscriptions = subscriptions
	}
}
