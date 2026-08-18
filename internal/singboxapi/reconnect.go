package singboxapi

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"

	"sing-scope/internal/domain"
)

// ReconnectOptions controls the retry and backoff behavior.
type ReconnectOptions struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	JitterFraction  float64
}

// DefaultReconnectOptions returns production backoff defaults.
func DefaultReconnectOptions() ReconnectOptions {
	return ReconnectOptions{
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     8 * time.Second,
		Multiplier:      1.5,
		JitterFraction:  0.2,
	}
}

// NextInterval calculates the next backoff duration with jitter.
func (o ReconnectOptions) NextInterval(current time.Duration) time.Duration {
	if current <= 0 {
		current = o.InitialInterval
	} else {
		current = time.Duration(float64(current) * o.Multiplier)
		if current > o.MaxInterval {
			current = o.MaxInterval
		}
	}

	jitter := float64(current) * o.JitterFraction * (rand.Float64()*2 - 1)
	result := time.Duration(float64(current) + jitter)
	if result < 100*time.Millisecond {
		result = 100 * time.Millisecond
	}
	return result
}

// Manager supervises the lifecycle of the sing-box API connection and its streams.
type Manager struct {
	mu             sync.RWMutex
	clientOpts     ClientOptions
	reconnectOpts  ReconnectOptions
	activeClient   *Client
	state          domain.ConnectionState
	serverVersion  string
	apiVersion     int32
	lastError      string
	connectedAt    *time.Time
	lastEventAt    *time.Time
	onStateChange  func(info *domain.ServerConnectionInfo)
	onBatch        ConnectionBatchHandler
	onStatus       StatusHandler
	onLogs         LogHandler
	onGroups       GroupsHandler
	cancelLoop     context.CancelFunc
	running        bool
}

// NewManager creates a connection manager.
func NewManager(
	clientOpts ClientOptions,
	reconnectOpts ReconnectOptions,
	onStateChange func(info *domain.ServerConnectionInfo),
	onBatch ConnectionBatchHandler,
	onStatus StatusHandler,
	onLogs LogHandler,
	onGroups GroupsHandler,
) *Manager {
	return &Manager{
		clientOpts:    clientOpts,
		reconnectOpts: reconnectOpts,
		state:         domain.StateDisconnected,
		onStateChange: onStateChange,
		onBatch:       onBatch,
		onStatus:      onStatus,
		onLogs:        onLogs,
		onGroups:      onGroups,
	}
}

// UpdateConfig updates target URL and secret, triggering a reconnect if already running.
func (m *Manager) UpdateConfig(serverURL, secret string) {
	m.mu.Lock()
	m.clientOpts.ServerURL = serverURL
	m.clientOpts.Secret = secret
	isRunning := m.running
	m.mu.Unlock()

	if isRunning {
		m.Restart()
	}
}

// GetInfo returns the current connection state snapshot.
func (m *Manager) GetInfo() *domain.ServerConnectionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &domain.ServerConnectionInfo{
		State:          m.state,
		ServerURL:      m.clientOpts.ServerURL,
		SingBoxVersion: m.serverVersion,
		APIVersion:     m.apiVersion,
		ErrorMessage:   m.lastError,
		ConnectedAt:    m.connectedAt,
		LastEventAt:    m.lastEventAt,
	}
}

func (m *Manager) setState(state domain.ConnectionState, errStr string) {
	m.mu.Lock()
	m.state = state
	m.lastError = errStr
	if state == domain.StateConnected {
		now := time.Now()
		m.connectedAt = &now
	}
	info := &domain.ServerConnectionInfo{
		State:          m.state,
		ServerURL:      m.clientOpts.ServerURL,
		SingBoxVersion: m.serverVersion,
		APIVersion:     m.apiVersion,
		ErrorMessage:   m.lastError,
		ConnectedAt:    m.connectedAt,
		LastEventAt:    m.lastEventAt,
	}
	cb := m.onStateChange
	m.mu.Unlock()

	if cb != nil {
		cb(info)
	}
}

// Start launches the supervisor loop in the background.
func (m *Manager) Start(parentCtx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(parentCtx)
	m.cancelLoop = cancel
	m.running = true
	m.mu.Unlock()

	go m.runLoop(ctx)
}

// Stop halts the supervisor and tears down any active connection.
func (m *Manager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	if m.cancelLoop != nil {
		m.cancelLoop()
		m.cancelLoop = nil
	}
	if m.activeClient != nil {
		_ = m.activeClient.Close()
		m.activeClient = nil
	}
	m.running = false
	m.mu.Unlock()

	m.setState(domain.StateDisconnected, "")
}

// Restart restarts the connection supervisor.
func (m *Manager) Restart() {
	m.Stop()
	m.Start(context.Background())
}

// CloseConnection proxies connection close requests to the active client.
func (m *Manager) CloseConnection(ctx context.Context, id string) error {
	m.mu.RLock()
	client := m.activeClient
	m.mu.RUnlock()

	if client == nil {
		return nil
	}
	return client.CloseConnection(ctx, id)
}

// CloseAllConnections proxies close-all requests to the active client.
func (m *Manager) CloseAllConnections(ctx context.Context) error {
	m.mu.RLock()
	client := m.activeClient
	m.mu.RUnlock()

	if client == nil {
		return nil
	}
	return client.CloseAllConnections(ctx)
}

// SelectOutbound proxies group outbound selection to the active client.
func (m *Manager) SelectOutbound(ctx context.Context, groupTag, outboundTag string) error {
	m.mu.RLock()
	client := m.activeClient
	m.mu.RUnlock()

	if client == nil {
		return nil
	}
	return client.SelectOutbound(ctx, groupTag, outboundTag)
}

// URLTest proxies latency testing to the active client.
func (m *Manager) URLTest(ctx context.Context, outboundTag string) error {
	m.mu.RLock()
	client := m.activeClient
	m.mu.RUnlock()

	if client == nil {
		return nil
	}
	return client.URLTest(ctx, outboundTag)
}

func (m *Manager) runLoop(ctx context.Context) {
	var backoff time.Duration
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		m.setState(domain.StateConnecting, "")

		m.mu.RLock()
		opts := m.clientOpts
		m.mu.RUnlock()

		client, err := NewClient(opts)
		if err != nil {
			m.setState(domain.StateError, err.Error())
			backoff = m.reconnectOpts.NextInterval(backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				continue
			}
		}

		// Verify API version via GetVersion
		info, err := client.GetVersion(ctx)
		if err != nil {
			_ = client.Close()
			m.setState(domain.StateError, err.Error())
			backoff = m.reconnectOpts.NextInterval(backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				continue
			}
		}

		m.mu.Lock()
		m.serverVersion = info.SingBoxVersion
		m.apiVersion = info.APIVersion
		m.activeClient = client
		m.mu.Unlock()

		if info.State == domain.StateIncompatible {
			_ = client.Close()
			m.setState(domain.StateIncompatible, info.ErrorMessage)
			backoff = m.reconnectOpts.NextInterval(backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				continue
			}
		}

		m.setState(domain.StateConnected, "")
		backoff = 0 // reset backoff on successful connect

		// Launch child streams
		streamCtx, cancelStreams := context.WithCancel(ctx)
		var wg sync.WaitGroup

		// Connection stream
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := RunConnectionStream(streamCtx, client, time.Second, func(events []domain.FlowEvent, isReset bool) {
				now := time.Now()
				m.mu.Lock()
				m.lastEventAt = &now
				m.mu.Unlock()

				if m.onBatch != nil {
					m.onBatch(events, isReset)
				}
			})
			if err != nil && streamCtx.Err() == nil {
				cancelStreams()
			}
		}()

		// Status stream
		if m.onStatus != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = RunStatusStream(streamCtx, client, time.Second, m.onStatus)
			}()
		}

		// Logs stream
		if m.onLogs != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = RunLogStream(streamCtx, client, m.onLogs)
			}()
		}

		// Groups stream
		if m.onGroups != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = RunGroupsStream(streamCtx, client, m.onGroups)
			}()
		}

		// Wait until streams exit or context cancelled
		wg.Wait()
		cancelStreams()
		_ = client.Close()

		m.mu.Lock()
		m.activeClient = nil
		m.mu.Unlock()

		if ctx.Err() != nil {
			return
		}

		m.setState(domain.StateReconnecting, "Connection lost, attempting reconnection...")
		backoff = m.reconnectOpts.NextInterval(backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
	}
}
