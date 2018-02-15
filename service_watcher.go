package main

import (
	"github.com/cameliot/alpaca"
	"github.com/cameliot/alpaca/meta"

	"log"
	"strings"
	"time"
)

type Manifest struct {
	Name        string    `json:"name"`
	Handle      string    `json:"handle"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	StartedAt   time.Time `json:"started_at"`
}

type GroupReport struct {
	LastRequest time.Time `json:"last_request"`
	LastSuccess time.Time `json:"last_success"`
	LastError   time.Time `json:"last_error"`

	TimeToReponse int64 `json:"time_to_response_us"`
}

type ServiceReport struct {
	Manifest      Manifest               `json:"manifest"`
	LastHeartbeat time.Time              `json:"last_heartbeat"`
	Groups        map[string]GroupReport `json:"groups"`
}

type WatchGroup struct {
	config *WatchConfig

	lastRequest time.Time
	lastSuccess time.Time
	lastError   time.Time

	lastResponse time.Time
	responseTime time.Duration
}

/*
Update times matching the configured action mapping
*/
func (self *WatchGroup) Update(handle string) {

	// Update times
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

	// Update response time:
	// Only recalculate response time if the last
	// Success or Error was before a request.
	//
	// This prevents updating the response time with an incorrect
	// (as in too high value) if another Success / Error was dispatched
	// not as a direct response, but as an 'event' generated by
	// external input or the service.
	if self.lastRequest.After(self.lastResponse) {
		// Update last response time
		lastResponse := self.lastSuccess
		if self.lastError.After(self.lastSuccess) {
			lastResponse = self.lastError
		}

		// Calculate response time
		self.responseTime = lastResponse.Sub(self.lastRequest)
		self.lastResponse = lastResponse
	}
}

type ServiceWatcher struct {
	config *ServiceConfig

	iama          meta.IamaPayload
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
			dispatch(meta.Whois(config.Handle))
			time.Sleep(1 * time.Minute)
		}
	}()

	return watcher
}

/*
Handle incoming actions and update watch groups
*/
func (self *ServiceWatcher) Handle(action alpaca.Action) {
	if action.Type == meta.PONG {
		self.handlePong(action)
		return
	} else if action.Type == meta.IAMA {
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
	pong := meta.DecodePong(action)
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
	iama := meta.DecodeIama(action)
	if iama.Handle != self.config.Handle {
		return // Not our concern
	}

	log.Println("Received Service Manifest for:", iama.Handle)
	self.iama = iama
}

/*
Generate service report
*/
func (self *ServiceWatcher) Report() ServiceReport {
	groupsReport := map[string]GroupReport{}
	for name, group := range self.watchGroups {

		// Make report
		groupsReport[name] = GroupReport{
			LastRequest:   group.lastRequest,
			LastSuccess:   group.lastSuccess,
			LastError:     group.lastError,
			TimeToReponse: int64(group.responseTime) / 1000,
		}
	}

	// Decode iama
	manifest := Manifest{
		Name:        self.iama.Name,
		Handle:      self.iama.Handle,
		Version:     self.iama.Version,
		Description: self.iama.Description,
		StartedAt:   self.iama.StartedAt(),
	}

	report := ServiceReport{
		LastHeartbeat: self.lastHeartbeat,
		Manifest:      manifest,
		Groups:        groupsReport,
	}

	return report
}
