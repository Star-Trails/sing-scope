package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"sing-scope/internal/app"
	"sing-scope/internal/domain"
	"sing-scope/internal/singboxapi"
	pb "sing-scope/internal/singboxapi/gen"
	"sing-scope/internal/store"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// mockStartedServiceServer implements pb.StartedServiceServer for full smoke testing.
type mockStartedServiceServer struct {
	pb.UnimplementedStartedServiceServer
	secret string
}

func (s *mockStartedServiceServer) GetVersion(ctx context.Context, _ *emptypb.Empty) (*pb.Version, error) {
	if err := s.checkAuth(ctx); err != nil {
		return nil, err
	}
	return &pb.Version{
		Version:    "1.14.0-beta.17",
		ApiVersion: 4,
	}, nil
}

func (s *mockStartedServiceServer) checkAuth(ctx context.Context) error {
	if s.secret == "" {
		return nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 || vals[0] != "Bearer "+s.secret {
		return fmt.Errorf("unauthenticated")
	}
	return nil
}

func (s *mockStartedServiceServer) SubscribeConnections(req *pb.SubscribeConnectionsRequest, stream grpc.ServerStreamingServer[pb.ConnectionEvents]) error {
	if err := s.checkAuth(stream.Context()); err != nil {
		return err
	}

	// 1. Initial snapshot batch with reset: true
	initialEvents := []*pb.ConnectionEvent{
		{
			Type: pb.ConnectionEventType_CONNECTION_EVENT_NEW,
			Id:   "flow-1111-aaaa",
			Connection: &pb.Connection{
				Id:            "flow-1111-aaaa",
				Inbound:       "tun-in",
				InboundType:   "tun",
				IpVersion:     4,
				Network:       "tcp",
				Source:        "172.19.0.1:52134",
				Destination:   "1.1.1.1:443",
				Domain:        "cloudflare.com",
				Protocol:      "tls",
				Outbound:      "HK 01",
				OutboundType:  "vless",
				Rule:          "geosite(category-games)",
				ChainList:     []string{"Proxy", "HK 01"},
				UplinkTotal:   102400,
				DownlinkTotal: 5242880,
				ProcessInfo: &pb.ProcessInfo{
					ProcessId:   1234,
					ProcessPath: "C:\\Program Files\\Google\\Chrome\\chrome.exe",
				},
				CreatedAt: time.Now().Add(-2 * time.Minute).UnixMilli(),
			},
		},
		{
			Type: pb.ConnectionEventType_CONNECTION_EVENT_NEW,
			Id:   "flow-2222-bbbb",
			Connection: &pb.Connection{
				Id:            "flow-2222-bbbb",
				Inbound:       "tun-in",
				InboundType:   "tun",
				IpVersion:     4,
				Network:       "udp",
				Source:        "172.19.0.1:52135",
				Destination:   "8.8.8.8:53",
				Domain:        "dns.google",
				Protocol:      "dns",
				Outbound:      "dns-out",
				OutboundType:  "direct",
				Rule:          "protocol: dns",
				ChainList:     []string{"dns-out"},
				UplinkTotal:   512,
				DownlinkTotal: 1024,
				ProcessInfo: &pb.ProcessInfo{
					ProcessId:   4567,
					ProcessPath: "/usr/bin/curl",
				},
				CreatedAt: time.Now().Add(-30 * time.Second).UnixMilli(),
			},
		},
	}

	_ = stream.Send(&pb.ConnectionEvents{
		Events: initialEvents,
		Reset_: true,
	})

	// 2. Send an update event with live rates
	time.Sleep(100 * time.Millisecond)
	_ = stream.Send(&pb.ConnectionEvents{
		Events: []*pb.ConnectionEvent{
			{
				Type:          pb.ConnectionEventType_CONNECTION_EVENT_UPDATE,
				Id:            "flow-1111-aaaa",
				UplinkDelta:   250000,
				DownlinkDelta: 2500000,
			},
		},
	})

	<-stream.Context().Done()
	return stream.Context().Err()
}

func (s *mockStartedServiceServer) SubscribeStatus(req *pb.SubscribeStatusRequest, stream grpc.ServerStreamingServer[pb.Status]) error {
	if err := s.checkAuth(stream.Context()); err != nil {
		return err
	}

	_ = stream.Send(&pb.Status{
		Memory:           48 * 1024 * 1024,
		Goroutines:       42,
		ConnectionsIn:    10,
		ConnectionsOut:   10,
		TrafficAvailable: true,
		Uplink:           250000,
		Downlink:         2500000,
		UplinkTotal:      1024000,
		DownlinkTotal:    52428800,
	})

	<-stream.Context().Done()
	return stream.Context().Err()
}

func (s *mockStartedServiceServer) SubscribeGroups(_ *emptypb.Empty, stream grpc.ServerStreamingServer[pb.Groups]) error {
	if err := s.checkAuth(stream.Context()); err != nil {
		return err
	}

	_ = stream.Send(&pb.Groups{
		Group: []*pb.Group{
			{
				Tag:        "Proxy",
				Type:       "selector",
				Selectable: true,
				Selected:   "HK 01",
				Items: []*pb.GroupItem{
					{Tag: "HK 01", Type: "vless", UrlTestDelay: 68, UrlTestTime: time.Now().Unix()},
					{Tag: "JP 02", Type: "hysteria2", UrlTestDelay: 125, UrlTestTime: time.Now().Unix()},
					{Tag: "US 03", Type: "shadowsocks", UrlTestDelay: 210, UrlTestTime: time.Now().Unix()},
				},
			},
		},
	})

	<-stream.Context().Done()
	return stream.Context().Err()
}

func (s *mockStartedServiceServer) SubscribeLog(_ *emptypb.Empty, stream grpc.ServerStreamingServer[pb.Log]) error {
	if err := s.checkAuth(stream.Context()); err != nil {
		return err
	}

	_ = stream.Send(&pb.Log{
		Messages: []*pb.Log_Message{
			{Level: pb.LogLevel_INFO, Message: "sing-box StartedService initialized on 127.0.0.1"},
			{Level: pb.LogLevel_INFO, Message: "[TUN] route: connection matched geosite(category-games) -> Proxy (HK 01)"},
		},
	})

	<-stream.Context().Done()
	return stream.Context().Err()
}

func (s *mockStartedServiceServer) SelectOutbound(ctx context.Context, req *pb.SelectOutboundRequest) (*emptypb.Empty, error) {
	if err := s.checkAuth(ctx); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *mockStartedServiceServer) URLTest(ctx context.Context, req *pb.URLTestRequest) (*emptypb.Empty, error) {
	if err := s.checkAuth(ctx); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *mockStartedServiceServer) CloseConnection(ctx context.Context, req *pb.CloseConnectionRequest) (*emptypb.Empty, error) {
	if err := s.checkAuth(ctx); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *mockStartedServiceServer) CloseAllConnections(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.checkAuth(ctx); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *mockStartedServiceServer) GetStartedAt(ctx context.Context, _ *emptypb.Empty) (*pb.StartedAt, error) {
	if err := s.checkAuth(ctx); err != nil {
		return nil, err
	}
	return &pb.StartedAt{StartedAt: time.Now().Add(-10 * time.Minute).UnixMilli()}, nil
}

func (s *mockStartedServiceServer) GetClashModeStatus(ctx context.Context, _ *emptypb.Empty) (*pb.ClashModeStatus, error) {
	if err := s.checkAuth(ctx); err != nil {
		return nil, err
	}
	return &pb.ClashModeStatus{CurrentMode: "rule", ModeList: []string{"rule", "global", "direct"}}, nil
}

func (s *mockStartedServiceServer) SetClashMode(ctx context.Context, req *pb.ClashMode) (*emptypb.Empty, error) {
	if err := s.checkAuth(ctx); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func main() {
	fmt.Println("=================================================================")
	fmt.Println("       sing-scope Full End-to-End System Smoke Test              ")
	fmt.Println("=================================================================")

	// 1. Start real in-process gRPC Server on local port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAILED: cannot listen: %v\n", err)
		os.Exit(1)
	}
	grpcPort := listener.Addr().(*net.TCPAddr).Port
	grpcURL := "http://" + listener.Addr().String()
	secret := "smoke-secret-test-token"

	grpcServer := grpc.NewServer()
	pb.RegisterStartedServiceServer(grpcServer, &mockStartedServiceServer{secret: secret})

	go func() {
		_ = grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()

	fmt.Printf("[1/6] Started mock sing-box gRPC StartedService on port %d\n", grpcPort)

	// 2. Initialize Go AppService & Manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connStore := store.NewConnectionStore(store.DefaultStoreOptions())
	connStore.SetInboundFilter("tun:all")

	mgr := singboxapi.NewManager(
		singboxapi.ClientOptions{
			ServerURL: grpcURL,
			Secret:    secret,
			Timeout:   2 * time.Second,
		},
		singboxapi.DefaultReconnectOptions(),
		func(info *domain.ServerConnectionInfo) {
			fmt.Printf("      -> Connection State Change: %s (version: %s, apiVersion: %d)\n",
				info.State, info.SingBoxVersion, info.APIVersion)
		},
		func(events []domain.FlowEvent, isReset bool) {
			connStore.ProcessBatch(events, isReset)
		},
		nil,
		nil,
		nil,
	)

	mgr.Start(ctx)
	defer mgr.Stop()

	// Wait for connection and initial stream batch
	fmt.Println("[2/6] Waiting for gRPC connection and initial flow stream delivery...")
	time.Sleep(300 * time.Millisecond)

	// 3. Test AppService & in-process AssetHandler HTTP endpoints
	appService := app.NewAppService()
	defer appService.Close()
	appService.ConnectServer(grpcURL, secret)
	time.Sleep(300 * time.Millisecond)

	// Create test HTTP server serving AssetHandler
	ts := httptest.NewServer(app.NewAssetHandler(appService, os.DirFS("frontend/dist")))
	defer ts.Close()
	fmt.Printf("[3/6] In-process AssetHandler mounted at: %s\n", ts.URL)

	// 4. Test & Validate All Endpoints
	fmt.Println("[4/6] Executing and asserting all REST endpoints:")

	testEndpoint(ts.URL+"/api/connection", "GET", nil, func(body []byte) {
		assertContains(body, `"state":"Connected"`, "Connection State")
		assertContains(body, `"singBoxVersion":"1.14.0-beta.17"`, "sing-box Version")
		assertContains(body, `"apiVersion":4`, "API Version 4")
	})

	testEndpoint(ts.URL+"/api/flows", "POST", map[string]any{"activeOnly": false, "limit": 50}, func(body []byte) {
		assertContains(body, `cloudflare.com`, "Target Domain (cloudflare.com)")
		assertContains(body, `HK 01`, "Outbound Tag (HK 01)")
		assertContains(body, `geosite(category-games)`, "Matched Rule")
	})

	testEndpoint(ts.URL+"/api/overview?filter=tun:all", "GET", nil, func(body []byte) {
		assertContains(body, `"activeTunFlows":`, "Active TUN Flows counter")
		assertContains(body, `"sessionDownload":`, "Session Download")
	})

	testEndpoint(ts.URL+"/api/analytics?filter=tun:all&topN=10", "GET", nil, func(body []byte) {
		assertContains(body, `"byDomain":`, "Analytics byDomain")
		assertContains(body, `"byDestination":`, "Analytics byDestination")
		assertContains(body, `"byOutbound":`, "Analytics byOutbound")
		assertContains(body, `"byRule":`, "Analytics byRule")
	})
	testEndpoint(ts.URL+"/api/groups", "GET", nil, func(body []byte) {
		assertContains(body, `"tag":"Proxy"`, "Outbound Group Tag")
		assertContains(body, `"HK 01"`, "Outbound Node HK 01")
		assertContains(body, `"urlTestDelay":68`, "URLTest Delay 68ms")
	})

	testEndpoint(ts.URL+"/api/rules", "GET", nil, func(body []byte) {
		assertContains(body, `"payload":`, "Rules payload")
		assertContains(body, `"proxy":`, "Rules target proxy")
	})

	testEndpoint(ts.URL+"/api/status", "GET", nil, func(body []byte) {
		assertContains(body, `"memory":`, "Memory stats")
		assertContains(body, `"goroutines":`, "Goroutines stats")
	})

	testEndpoint(ts.URL+"/api/logs?limit=50", "GET", nil, func(body []byte) {
		assertContains(body, `StartedService initialized`, "Log Message")
	})

	testEndpoint(ts.URL+"/api/select-outbound", "POST", map[string]string{"groupTag": "Proxy", "outboundTag": "JP 02"}, func(body []byte) {
		assertContains(body, `"ok":true`, "SelectOutbound response")
	})

	testEndpoint(ts.URL+"/api/url-test?tag=Proxy", "POST", nil, func(body []byte) {
		assertContains(body, `"ok":true`, "URLTest response")
	})
	testEndpoint(ts.URL+"/api/probe-latency?target=baidu", "GET", nil, func(body []byte) {
		assertContains(body, `"latency":`, "ProbeLatency response")
	})


	// 5. Test Frontend Asset SPA Fallback
	fmt.Println("[5/6] Testing Frontend SPA Asset delivery:")
	testEndpoint(ts.URL+"/", "GET", nil, func(body []byte) {
		assertContains(body, `<div id="app"`, "SPA Root index.html")
	})
	testEndpoint(ts.URL+"/overview", "GET", nil, func(body []byte) {
		assertContains(body, `<div id="app"`, "SPA Fallback Route (/overview)")
	})

	// 6. Final Summary
	fmt.Println("[6/6] All Smoke Tests PASSED successfully!")
	fmt.Println("=================================================================")
	fmt.Println("  RESULT: SUCCESS (gRPC -> Go Store -> Rust Core -> HTTP Bridge) ")
	fmt.Println("=================================================================")
}

func testEndpoint(url, method string, reqBody any, validator func([]byte)) {
	var bodyReader io.Reader
	if reqBody != nil {
		b, _ := json.Marshal(reqBody)
		bodyReader = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(method, url, bodyReader)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  FAILED: %s %s error: %v\n", method, url, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "  FAILED: %s %s status code: %d\n", method, url, resp.StatusCode)
		os.Exit(1)
	}

	body, _ := io.ReadAll(resp.Body)
	validator(body)
	shortURL := url
	if idx := strings.Index(url, "/api"); idx != -1 {
		shortURL = url[idx:]
	}
	fmt.Printf("  [PASS] %-6s %-32s -> OK (%d bytes)\n", method, shortURL, len(body))
}

func assertContains(body []byte, substring, label string) {
	if !strings.Contains(string(body), substring) {
		fmt.Fprintf(os.Stderr, "  ASSERTION FAILED for [%s]: expected to contain %q, but got:\n%s\n",
			label, substring, string(body))
		os.Exit(1)
	}
}
