package app

import (
	"context"
	"sync"
	"time"

	"sing-scope/internal/analytics"
	"sing-scope/internal/domain"
	"sing-scope/internal/singboxapi"
	"sing-scope/internal/store"
)

// AppService is the centralized application service bound to the Wails desktop frontend.
type AppService struct {
	ctx              context.Context
	cancel           context.CancelFunc
	store            *store.ConnectionStore
	analyticsService *analytics.Service
	manager          *singboxapi.Manager

	mu           sync.RWMutex
	systemStatus *domain.SystemStatus
	logs         []domain.LogMessage
	groups       []domain.OutboundGroup
	maxLogs      int
}

// NewAppService creates and initializes the core application service.
func NewAppService() *AppService {
	ctx, cancel := context.WithCancel(context.Background())
	connStore := store.NewConnectionStore(store.DefaultStoreOptions())
	analyticsSvc := analytics.NewService(connStore)

	svc := &AppService{
		ctx:              ctx,
		cancel:           cancel,
		store:            connStore,
		analyticsService: analyticsSvc,
		logs:             make([]domain.LogMessage, 0, 500),
		maxLogs:          500,
	}

	mgr := singboxapi.NewManager(
		singboxapi.ClientOptions{
			ServerURL: "http://127.0.0.1:9090",
			Secret:    "",
			Timeout:   5 * time.Second,
		},
		singboxapi.DefaultReconnectOptions(),
		func(info *domain.ServerConnectionInfo) {
			// Callback on state change
		},
		func(events []domain.FlowEvent, isReset bool) {
			connStore.ProcessBatch(events, isReset)
		},
		func(status *domain.SystemStatus) {
			svc.mu.Lock()
			svc.systemStatus = status
			svc.mu.Unlock()
		},
		func(logs []domain.LogMessage) {
			svc.mu.Lock()
			for _, l := range logs {
				if len(svc.logs) >= svc.maxLogs {
					svc.logs = append(svc.logs[1:], l)
				} else {
					svc.logs = append(svc.logs, l)
				}
			}
			svc.mu.Unlock()
		},
		func(groups []domain.OutboundGroup) {
			svc.mu.Lock()
			svc.groups = groups
			svc.mu.Unlock()
		},
	)

	svc.manager = mgr
	mgr.Start(ctx)

	return svc
}

// GetConnectionInfo returns the active connection details and version.
func (s *AppService) GetConnectionInfo() domain.ServerConnectionInfo {
	return *s.manager.GetInfo()
}

// ConnectServer connects to a specified sing-box API server address.
func (s *AppService) ConnectServer(url, secret string) bool {
	s.manager.UpdateConfig(url, secret)
	return true
}

// DisconnectServer stops connection to the current server.
func (s *AppService) DisconnectServer() bool {
	s.manager.Stop()
	return true
}

// GetOverviewSummary returns the live traffic rate chart data and KPI cards.
func (s *AppService) GetOverviewSummary(inboundFilter string) domain.OverviewSummary {
	if inboundFilter != "" {
		s.store.SetInboundFilter(inboundFilter)
	}
	return s.store.GetOverviewSummary()
}

// GetFlows returns paginated and filtered network flow records.
func (s *AppService) GetFlows(opts store.QueryOptions) store.FlowListResult {
	return s.store.GetFlows(opts)
}

// GetBatchAnalytics computes multi-dimensional breakdowns via Rust/Go engine.
func (s *AppService) GetBatchAnalytics(inboundFilter string, topN int) analytics.BatchAnalysisResult {
	res, err := s.analyticsService.AnalyzeBatch(inboundFilter, topN)
	if err != nil {
		return analytics.BatchAnalysisResult{Engine: "error: " + err.Error()}
	}
	return res
}

// GetSystemStatus returns the latest sing-box process metrics (memory, goroutines).
func (s *AppService) GetSystemStatus() domain.SystemStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.systemStatus != nil {
		return *s.systemStatus
	}
	return domain.SystemStatus{
		Timestamp: time.Now(),
	}
}

// GetLogs returns recent sing-box log messages.
func (s *AppService) GetLogs(limit int) []domain.LogMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.logs) {
		limit = len(s.logs)
	}

	res := make([]domain.LogMessage, limit)
	copy(res, s.logs[len(s.logs)-limit:])
	return res
}

// GetGroups returns current outbound selector groups.
func (s *AppService) GetGroups() []domain.OutboundGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]domain.OutboundGroup, len(s.groups))
	copy(res, s.groups)
	return res
}

// CloseConnection closes a specific network connection.
func (s *AppService) CloseConnection(id string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.manager.CloseConnection(ctx, id)
	return err == nil
}

// CloseAllConnections terminates all connections tracked by sing-box.
func (s *AppService) CloseAllConnections() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.manager.CloseAllConnections(ctx)
	return err == nil
}
// SelectOutbound switches the selected node in an outbound selector group.
func (s *AppService) SelectOutbound(groupTag, outboundTag string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.manager.SelectOutbound(ctx, groupTag, outboundTag)
	return err == nil
}

// URLTest triggers an outbound latency test.
func (s *AppService) URLTest(outboundTag string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	err := s.manager.URLTest(ctx, outboundTag)
	return err == nil
}

// SetInboundFilter updates the active TUN filter.
func (s *AppService) SetInboundFilter(filter string) bool {
	s.store.SetInboundFilter(filter)
	return true
}

// Close cleans up background loops and active streams.
func (s *AppService) Close() {
	s.cancel()
	s.manager.Stop()
}
