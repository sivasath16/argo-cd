package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	"github.com/argoproj/argo-cd/v3/common"
)

// LoggerRecoveryHandler return a handler for recovering from panics and returning error
func LoggerRecoveryHandler(log *logrus.Entry) recovery.RecoveryHandlerFunc {
	return func(p any) (err error) {
		log.Errorf("Recovered from panic: %+v\n%s", p, debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}
}

// BlockingDial is a helper method to dial the given address, using optional TLS credentials,
// and blocking until the returned connection is ready. If the given credentials are nil, the
// connection will be insecure (plain-text).
// Lifted from: https://github.com/fullstorydev/grpcurl/blob/master/grpcurl.go
func BlockingDial(ctx context.Context, network, address string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// grpc.Dial doesn't provide any information on permanent connection errors (like
	// TLS handshake failures). So in order to provide good error messages, we need a
	// custom dialer that can provide that info. That means we manage the TLS handshake.
	result := make(chan any, 1)
	writeResult := func(res any) {
		// non-blocking write: we only need the first result
		select {
		case result <- res:
		default:
		}
	}

	dialer := func(ctx context.Context, address string) (net.Conn, error) {
		conn, err := proxy.Dial(ctx, network, address)
		if err != nil {
			writeResult(err)
			return nil, fmt.Errorf("error dial proxy: %w", err)
		}
		if creds != nil {
			conn, _, err = creds.ClientHandshake(ctx, address, conn)
			if err != nil {
				writeResult(err)
				return nil, fmt.Errorf("error creating connection: %w", err)
			}
		}
		return conn, nil
	}

	// Even with grpc.FailOnNonTempDialError, this call will usually timeout in
	// the face of TLS handshake errors. So we can't rely on grpc.WithBlock() to
	// know when we're done. So we run it in a goroutine and then use result
	// channel to either get the channel or fail-fast.
	go func() {
		opts = append(opts,
			//nolint:staticcheck
			grpc.WithBlock(),
			//nolint:staticcheck
			grpc.FailOnNonTempDialError(true),
			grpc.WithContextDialer(dialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()), // we are handling TLS, so tell grpc not to
			grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: common.GetGRPCKeepAliveTime()}),
		)
		//nolint:staticcheck
		conn, err := grpc.DialContext(ctx, address, opts...)
		var res any
		if err != nil {
			res = err
		} else {
			res = conn
		}
		writeResult(res)
	}()

	select {
	case res := <-result:
		if conn, ok := res.(*grpc.ClientConn); ok {
			return conn, nil
		}
		return nil, res.(error)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type TLSTestResult struct {
	TLS         bool
	InsecureErr error
}

func TestTLS(address string, dialTime time.Duration) (*TLSTestResult, error) {
	if parts := strings.Split(address, ":"); len(parts) == 1 {
		// If port is unspecified, assume the most likely port
		address += ":443"
	}
	var testResult TLSTestResult
	var tlsConfig tls.Config
	tlsConfig.InsecureSkipVerify = true
	creds := credentials.NewTLS(&tlsConfig)

	// Set timeout when dialing to the server
	// fix: https://github.com/argoproj/argo-cd/issues/9679
	ctx, cancel := context.WithTimeout(context.Background(), dialTime)
	defer cancel()

	conn, err := BlockingDial(ctx, "tcp", address, creds)
	if err == nil {
		_ = conn.Close()
		testResult.TLS = true
		creds := credentials.NewTLS(&tls.Config{})
		ctx, cancel := context.WithTimeout(context.Background(), dialTime)
		defer cancel()

		conn, err := BlockingDial(ctx, "tcp", address, creds)
		if err == nil {
			_ = conn.Close()
		} else {
			// if connection was successful with InsecureSkipVerify true, but unsuccessful with
			// InsecureSkipVerify false, it means server is not configured securely
			testResult.InsecureErr = err
		}
		return &testResult, nil
	}
	// If we get here, we were unable to connect via TLS (even with InsecureSkipVerify: true)
	// It may be because server is running without TLS, or because of real issues (e.g. connection
	// refused). Test if server accepts plain-text connections
	ctx, cancel = context.WithTimeout(context.Background(), dialTime)
	defer cancel()
	conn, err = BlockingDial(ctx, "tcp", address, nil)
	if err == nil {
		_ = conn.Close()
		testResult.TLS = false
		return &testResult, nil
	}
	return nil, err
}
