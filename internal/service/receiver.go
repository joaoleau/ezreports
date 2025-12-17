package service

import (
	"github.com/joaoleau/ezreports/internal/config"
)

func GetReceiver (receiverName string, receiver []config.Receiver) (*config.Receiver) {	
	for i := range receiver {
		if receiverName == receiver[i].Name {
			return &receiver[i]
		}
	}

	return nil
}