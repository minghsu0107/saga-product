package grpc

import (
	"context"
	"fmt"
	"time"

	"go.opencensus.io/plugin/ocgrpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

func InitializeServer(ocAgentHost string, logrusEntry *log.Entry) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 8), // increase to 8 MB (default: 4 MB)
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // terminate the connection if a client pings more than once every 5 seconds
			PermitWithoutStream: true,            // allow pings even when there are no active streams
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,  // if a client is idle for 15 seconds, send a GOAWAY
			MaxConnectionAge:      600 * time.Second, // if any connection is alive for more than maxConnectionAge, send a GOAWAY
			MaxConnectionAgeGrace: 5 * time.Second,   // allow 5 seconds for pending RPCs to complete before forcibly closing connections
			Time:                  5 * time.Second,   // ping the client if it is idle for 5 seconds to ensure the connection is still active
			Timeout:               1 * time.Second,   // wait 1 second for the ping ack before assuming the connection is dead
		}),
	}
	if ocAgentHost != "" {
		opts = append(opts, grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	}

	grpc_prometheus.EnableHandlingTimeHistogram()

	recoveryFunc := func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoveryFunc),
	}
	grpcOpts := []grpc_logrus.Option{
		grpc_logrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}
	//grpc_logrus.ReplaceGrpcLogger(logrusEntry)

	opts = append(opts,
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_prometheus.StreamServerInterceptor,
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(logrusEntry, grpcOpts...),
			grpc_recovery.StreamServerInterceptor(recoveryOpts...),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.UnaryServerInterceptor(logrusEntry, grpcOpts...),
			grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
		)),
	)
	return grpc.NewServer(opts...)
}

func InitializeClient(svcHost, ocAgentHost string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	scheme := "dns"

	retryOpts := []grpc_retry.CallOption{
		// generate waits between 900ms to 1100ms
		grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(1*time.Second, 0.1)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
	}

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	if ocAgentHost != "" {
		dialOpts = append(dialOpts, grpc.WithStatsHandler(new(ocgrpc.ClientHandler)))
	}

	dialOpts = append(dialOpts,
		grpc.WithDisableServiceConfig(),
		grpc.WithDefaultServiceConfig(`{
			"loadBalancingPolicy": "round_robin"
		}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		}),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
		//grpc.WithBlock(),
	)

	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:///%s", scheme, svcHost),
		dialOpts...,
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
