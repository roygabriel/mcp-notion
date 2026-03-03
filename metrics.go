package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	toolCallsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mcp_tool_calls_total",
		Help: "Total number of MCP tool calls.",
	}, []string{"tool", "status"})

	toolCallDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "mcp_tool_call_duration_seconds",
		Help:    "Duration of MCP tool calls in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"tool"})
)
