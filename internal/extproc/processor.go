// Package extproc implements the external processing filter handler
// for Envoy's ext_proc gRPC service. It intercepts HTTP requests and
// responses to apply AI gateway transformations such as routing,
// rate limiting, and provider-specific request/response normalization.
package extproc

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Processor handles the ext_proc bidirectional streaming RPC.
// It processes request and response phases for each HTTP transaction.
type Processor struct {
	extprocv3.UnimplementedExternalProcessorServer
	logger *slog.Logger
}

// NewProcessor creates a new Processor instance with the given logger.
func NewProcessor(logger *slog.Logger) *Processor {
	if logger == nil {
		logger = slog.Default()
	}
	return &Processor{logger: logger}
}

// Process implements the ExternalProcessorServer interface.
// It handles the bidirectional stream of processing messages from Envoy.
func (p *Processor) Process(stream extprocv3.ExternalProcessor_ProcessServer) error {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "failed to receive request: %v", err)
		}

		resp, err := p.handleRequest(ctx, req)
		if err != nil {
			p.logger.ErrorContext(ctx, "error handling ext_proc request", "error", err)
			return status.Errorf(codes.Internal, "processing error: %v", err)
		}

		if err := stream.Send(resp); err != nil {
			return status.Errorf(codes.Unknown, "failed to send response: %v", err)
		}
	}
}

// handleRequest dispatches the incoming ProcessingRequest to the appropriate
// phase handler based on which phase is present in the request.
func (p *Processor) handleRequest(ctx context.Context, req *extprocv3.ProcessingRequest) (*extprocv3.ProcessingResponse, error) {
	switch r := req.Request.(type) {
	case *extprocv3.ProcessingRequest_RequestHeaders:
		return p.handleRequestHeaders(ctx, r.RequestHeaders)
	case *extprocv3.ProcessingRequest_RequestBody:
		return p.handleRequestBody(ctx, r.RequestBody)
	case *extprocv3.ProcessingRequest_ResponseHeaders:
		return p.handleResponseHeaders(ctx, r.ResponseHeaders)
	case *extprocv3.ProcessingRequest_ResponseBody:
		return p.handleResponseBody(ctx, r.ResponseBody)
	default:
		return nil, fmt.Errorf("unknown request type: %T", req.Request)
	}
}

// handleRequestHeaders processes incoming HTTP request headers.
// This is where AI provider routing decisions are made.
func (p *Processor) handleRequestHeaders(ctx context.Context, headers *extprocv3.HttpHeaders) (*extprocv3.ProcessingResponse, error) {
	p.logger.DebugContext(ctx, "processing request headers")
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extprocv3.HeadersResponse{},
		},
	}, nil
}

// handleRequestBody processes the HTTP request body.
// This is where request transformation for AI providers occurs.
func (p *Processor) handleRequestBody(ctx context.Context, body *extprocv3.HttpBody) (*extprocv3.ProcessingResponse, error) {
	p.logger.DebugContext(ctx, "processing request body", "end_of_stream", body.EndOfStream)
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestBody{
			RequestBody: &extprocv3.BodyResponse{},
		},
	}, nil
}

// handleResponseHeaders processes the HTTP response headers from the upstream AI provider.
func (p *Processor) handleResponseHeaders(ctx context.Context, headers *extprocv3.HttpHeaders) (*extprocv3.ProcessingResponse, error) {
	p.logger.DebugContext(ctx, "processing response headers")
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ResponseHeaders{
			ResponseHeaders: &extprocv3.HeadersResponse{},
		},
	}, nil
}

// handleResponseBody processes the HTTP response body from the upstream AI provider.
// This is where response normalization and token usage tracking occurs.
func (p *Processor) handleResponseBody(ctx context.Context, body *extprocv3.HttpBody) (*extprocv3.ProcessingResponse, error) {
	p.logger.DebugContext(ctx, "processing response body", "end_of_stream", body.EndOfStream)
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ResponseBody{
			ResponseBody: &extprocv3.BodyResponse{},
		},
	}, nil
}
