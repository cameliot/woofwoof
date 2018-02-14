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

	TimeToReponse time.Duration `json:"time_to_response_us"`
}

type ServiceReport struct {
	Manifest      IamaPayload            `json:"manifest"`
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

	manifest      IamaPayload
	lastHeartbeat time.Time
	watchGroups   map[string]*WatchGroup
}

/*
Initialize new service watcher
*/
func NewServiceWatcher(config *ServiceConfig, dispatch alpaca.Dispatch) *ServiceWatcher {
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

	// Periodically request WHOIS information from
	// this service
	go func() {
		for {
			dispatch(Whois(config.Handle))
			time.Sleep(1 * time.Minute)
		}
	}()

	return watcher
}

/*
Handle incoming actions and update watch groups
*/
func (self *ServiceWatcher) Handle(action alpaca.Action) {
	if action.Type == PONG {
		self.handlePong(action)
		return
	} else if action.Type == IAMA {
		self.handleIama(action)
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
	pong := DecodePong(action)
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

func (self *ServiceWatcher) handleIama(action alpaca.Action) {
	// Decode Payload
	iama := DecodeIama(action)
	if iama.Handle != self.config.Handle {
		return // Not our concern
	}

	log.Println("Received Service Manifest for:", iama.Handle)
	self.manifest = iama
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
		// Use microseconds instead of nanoseconds,
		responseTime /= 1000

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
		Manifest:      self.manifest,
		Groups:        groupsReport,
	}

	return report
}
