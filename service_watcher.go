package main

import (
	"fmt"
	"github.com/cameliot/alpaca"
	"time"
)

type GroupReport struct {
	LastRequest time.Time `json:"last_request"`
	LastSuccess time.Time `json:"last_success"`
	LastError   time.Time `json:"last_error"`

	TimeToReponse time.Duration `json:"time_to_response"`
}

type ServiceReport struct {
	Handle string `json:"handle"`

	LastHeartbeat time.Time `json:"last_heartbeat"`

	Groups map[string]GroupReport `json:"groups"`
}

type WatchGroup struct {
	config *WatchConfig

	lastRequest time.Time
	lastSuccess time.Time
	lastError   time.Time
}

type ServiceWatcher struct {
	config *ServiceConfig

	lastHeartbeat time.Time
	watchGroups   map[string]WatchGroup
}

/*
Initialize new service watcher
*/
func NewServiceWatcher(config *ServiceConfig) *ServiceWatcher {
	// Create watch groups
	watchGroups := map[string]WatchGroup{}
	for handle, watchConfig := range config.Watches {
		watchGroups[handle] = WatchGroup{
			config: &watchConfig,
		}
	}

	watcher := &ServiceWatcher{
		config:      config,
		watchGroups: watchGroups,
	}

	return watcher
}

/*
Handle incoming actions and update watch groups
*/
func (self *ServiceWatcher) Handle(action alpaca.Action) {
	switch action.Type {
	case "@meta/PONG":
		self.handlePong(action)
	}
}

func (self *ServiceWatcher) handlePong(action alpaca.Action) {
	// Decode Payload
	fmt.Println("Handleing action:", action)
}

/*
Generate service report
*/
func (self *ServiceWatcher) Report() ServiceReport {
	groupsReport := map[string]GroupReport{}
	for name, group := range self.watchGroups {
		// Calculate response time
		lastResponse := group.lastSuccess
		if group.lastError.After(group.lastSuccess) {
			lastResponse = group.lastError
		}

		responseTime := lastResponse.Sub(group.lastRequest)

		// Make report
		groupsReport[name] = GroupReport{
			LastRequest:   group.lastRequest,
			LastSuccess:   group.lastSuccess,
			LastError:     group.lastError,
			TimeToReponse: responseTime,
		}
	}

	report := ServiceReport{
		Handle:        self.config.Handle,
		LastHeartbeat: self.lastHeartbeat,
		Groups:        groupsReport,
	}

	return report
}
