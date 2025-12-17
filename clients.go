package main

import (
	"sync"

	"go.mau.fi/whatsmeow"
)

type ClientManager struct {
	sync.RWMutex
	whatsmeowClient *whatsmeow.Client
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
