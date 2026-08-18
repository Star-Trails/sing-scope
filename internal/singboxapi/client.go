package singboxapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"time"

	"sing-scope/internal/domain"
	pb "sing-scope/internal/singboxapi/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ClientOptions configures the sing-box API gRPC client.
type ClientOptions struct {
	ServerURL string
	Secret    string
	Timeout   time.Duration
}

// Client is a strongly typed client for the new sing-box gRPC API.
type Client struct {
	conn       *grpc.ClientConn
	grpcClient pb.StartedServiceClient
	options    ClientOptions
	target     string
	isTLS      bool
}

// ParseTargetAndCreds parses an http/https server URL into a gRPC target and transport credentials.
func ParseTargetAndCreds(serverURL string) (string, credentials.TransportCredentials, bool, error) {
	if serverURL == "" {
		return "", nil, false, fmt.Errorf("empty server URL")
	}

	u, err := url.Parse(serverURL)
	if err != nil {
		return "", nil, false, fmt.Errorf("invalid server URL: %w", err)
	}

	var isTLS bool
	switch u.Scheme {
	case "http":
		isTLS = false
	case "https":
		isTLS = true
	default:
		return "", nil, false, fmt.Errorf("unsupported scheme %q (expected http or https)", u.Scheme)
	}

	host := u.Hostname()
	if host == "" {
		return "", nil, false, fmt.Errorf("missing hostname in server URL: %s", serverURL)
	}

	port := u.Port()
	if port == "" {
		if isTLS {
			port = "443"
		} else {
			port = "80"
		}
	}

	var creds credentials.TransportCredentials
	if isTLS {
		creds = credentials.NewTLS(&tls.Config{
			ServerName: host,
		})
	} else {
		creds = insecure.NewCredentials()
	}

	return net.JoinHostPort(host, port), creds, isTLS, nil
}

// NewClient creates and establishes a gRPC client connection to sing-box API service.
func NewClient(options ClientOptions) (*Client, error) {
	target, creds, isTLS, err := ParseTargetAndCreds(options.ServerURL)
	if err != nil {
		return nil, err
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithChainUnaryInterceptor(NewAuthUnaryInterceptor(options.Secret)),
		grpc.WithChainStreamInterceptor(NewAuthStreamInterceptor(options.Secret)),
	}

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &Client{
		conn:       conn,
		grpcClient: pb.NewStartedServiceClient(conn),
		options:    options,
		target:     target,
		isTLS:      isTLS,
	}, nil
}

// Conn returns the underlying gRPC ClientConn.
func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

// StartedServiceClient returns the generated proto client interface.
func (c *Client) StartedServiceClient() pb.StartedServiceClient {
	return c.grpcClient
}

// GetVersion queries the sing-box version and API version.
func (c *Client) GetVersion(ctx context.Context) (*domain.ServerConnectionInfo, error) {
	timeout := c.options.Timeout
	if timeout <= 0 {
		timeout = 1500 * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	v, err := c.grpcClient.GetVersion(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("GetVersion failed: %w", err)
	}

	now := time.Now()
	compat := CheckCompatibility(v.GetVersion(), v.GetApiVersion())

	state := domain.StateConnected
	if !compat.Compatible {
		state = domain.StateIncompatible
	}

	return &domain.ServerConnectionInfo{
		State:          state,
		ServerURL:      c.options.ServerURL,
		SingBoxVersion: v.GetVersion(),
		APIVersion:     v.GetApiVersion(),
		ErrorMessage:   compat.Message,
		ConnectedAt:    &now,
		LastEventAt:    &now,
	}, nil
}

// SubscribeConnections opens a server-streaming connection event subscription.
func (c *Client) SubscribeConnections(ctx context.Context, interval time.Duration) (grpc.ServerStreamingClient[pb.ConnectionEvents], error) {
	intervalNs := int64(interval)
	if intervalNs <= 0 {
		intervalNs = int64(time.Second)
	}
	return c.grpcClient.SubscribeConnections(ctx, &pb.SubscribeConnectionsRequest{
		Interval: intervalNs,
	})
}

// SubscribeStatus opens a server-streaming system status subscription.
func (c *Client) SubscribeStatus(ctx context.Context, interval time.Duration) (grpc.ServerStreamingClient[pb.Status], error) {
	intervalNs := int64(interval)
	if intervalNs <= 0 {
		intervalNs = int64(time.Second)
	}
	return c.grpcClient.SubscribeStatus(ctx, &pb.SubscribeStatusRequest{
		Interval: intervalNs,
	})
}

// SubscribeLog opens a server-streaming log subscription.
func (c *Client) SubscribeLog(ctx context.Context) (grpc.ServerStreamingClient[pb.Log], error) {
	return c.grpcClient.SubscribeLog(ctx, &emptypb.Empty{})
}

// SubscribeGroups opens a server-streaming outbound group subscription.
func (c *Client) SubscribeGroups(ctx context.Context) (grpc.ServerStreamingClient[pb.Groups], error) {
	return c.grpcClient.SubscribeGroups(ctx, &emptypb.Empty{})
}

// SubscribeOutbounds opens a server-streaming outbound list subscription.
func (c *Client) SubscribeOutbounds(ctx context.Context) (grpc.ServerStreamingClient[pb.OutboundList], error) {
	return c.grpcClient.SubscribeOutbounds(ctx, &emptypb.Empty{})
}

// CloseConnection closes an active flow by its UUID string ID.
func (c *Client) CloseConnection(ctx context.Context, id string) error {
	_, err := c.grpcClient.CloseConnection(ctx, &pb.CloseConnectionRequest{Id: id})
	return err
}

// CloseAllConnections closes all active connections in sing-box.
func (c *Client) CloseAllConnections(ctx context.Context) error {
	_, err := c.grpcClient.CloseAllConnections(ctx, &emptypb.Empty{})
	return err
}

// SelectOutbound selects an outbound inside a selector group.
func (c *Client) SelectOutbound(ctx context.Context, groupTag, outboundTag string) error {
	_, err := c.grpcClient.SelectOutbound(ctx, &pb.SelectOutboundRequest{
		GroupTag:    groupTag,
		OutboundTag: outboundTag,
	})
	return err
}

// URLTest triggers an outbound group latency URL test.
func (c *Client) URLTest(ctx context.Context, outboundTag string) error {
	_, err := c.grpcClient.URLTest(ctx, &pb.URLTestRequest{
		OutboundTag: outboundTag,
	})
	return err
}

// Close terminates the gRPC connection channel.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
