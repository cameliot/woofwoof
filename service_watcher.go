package main

import (
	"github.com/cameliot/alpaca"
	"log"
	"strings"
	"time"
)

type GroupReport struct {
	LastRequest time.Time `json:"last_request"`
	LastSuccess time.Time `json:"last_success"`
	LastError   time.Time `json:"last_error"`

	TimeToReponse time.Duration `json:"time_to_response"`
}

type ServiceReport struct {
	LastHeartbeat time.Time              `json:"last_heartbeat"`
	Groups        map[string]GroupReport `json:"groups"`
}

type WatchGroup struct {
	config *WatchConfig

	lastRequest time.Time
	lastSuccess time.Time
	lastError   time.Time
}

/*
Update times matching the configured action mapping
*/
func (self *WatchGroup) Update(handle string) {
	switch handle {
	case self.config.Request:
		self.lastRequest = time.Now().UTC()
		break
	case self.config.Success:
		self.lastSuccess = time.Now().UTC()
		break
	case self.config.Error:
		self.lastError = time.Now().UTC()
	}
}

type ServiceWatcher struct {
	config *ServiceConfig

	lastHeartbeat time.Time
	watchGroups   map[string]*WatchGroup
}

/*
Initialize new service watcher
*/
func NewServiceWatcher(config *ServiceConfig) *ServiceWatcher {
	// Create watch groups
	watchGroups := map[string]*WatchGroup{}
	for handle, watchConfig := range config.Watches {
		watchGroups[handle] = &WatchGroup{
			config: watchConfig,
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
	if action.Type == PONG {
		self.handlePong(action)
		return
	}

	// Handle all other incoming actions, matching
	// this service
	if !strings.HasPrefix(action.Type, "@"+self.config.Handle) {
		return // Not our concern
	}

	// Get action handle
	tokens := strings.Split(action.Type, "/")
	handle := tokens[len(tokens)-1]

	for _, group := range self.watchGroups {
		group.Update(handle)
	}
}

func (self *ServiceWatcher) handlePong(action alpaca.Action) {

	// Decode Payload
	pong := DecodePongPayload(action)
	if pong.Handle != self.config.Handle {
		return // Not our concern
	}

	// Update heartbeat
	self.lastHeartbeat = pong.Timestamp()
	log.Println(
		"Received Heartbeat for", pong.Handle,
		":", pong.Timestamp(),
	)
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
		LastHeartbeat: self.lastHeartbeat,
		Groups:        groupsReport,
	}

	return report
}
