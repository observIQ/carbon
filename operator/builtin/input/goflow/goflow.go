package goflow

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	flowmessage "github.com/cloudflare/goflow/v3/pb"
	"github.com/cloudflare/goflow/v3/utils"
	"github.com/jpillora/backoff"
	"github.com/observiq/stanza/operator"
	"github.com/observiq/stanza/operator/helper"
	"go.uber.org/zap"
)

const (
	operatorName  = "goflow_input"
	modeSflow     = "sflow"
	modeNetflowV5 = "netflow_v5"
	modeNetflowV9 = "netflow_v9"
)

func init() {
	operator.Register(operatorName, func() operator.Builder { return NewGoflowInputConfig("") })
}

// NewGoflowInputConfig creates a new goflow input config with default values
func NewGoflowInputConfig(operatorID string) *GoflowInputConfig {
	return &GoflowInputConfig{
		InputConfig: helper.NewInputConfig(operatorID, operatorName),
	}
}

// GoflowInputConfig is the configuration of a goflow input operator.
type GoflowInputConfig struct {
	helper.InputConfig `yaml:",inline"`

	Mode    string `json:"mode,omitempty"    yaml:"mode,omitempty"`
	Address string `json:"address,omitempty" yaml:"address,omitempty"`
	Port    uint   `json:"port,omitempty"    yaml:"port,omitempty"`
	Workers uint   `json:"workers,omitempty" yaml:"workers,omitempty"`
}

// Build will build a goflow input operator.
func (c *GoflowInputConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(context)
	if err != nil {
		return nil, err
	}

	switch c.Mode {
	case "sflow", "netflow_v5", "netflow_v9":
		break
	default:
		return nil, fmt.Errorf("%s is not a supported Goflow mode", c.Mode)
	}

	if c.Address == "" {
		c.Address = "0.0.0.0"
	}

	if c.Port == 0 {
		return nil, fmt.Errorf("port is a required parameter")
	}

	addr := net.UDPAddr{
		IP:   net.ParseIP(c.Address),
		Port: int(c.Port),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, fmt.Errorf("expected udp socket %s:%d to be available, got %w", c.Address, c.Port, err)
	}
	if err := conn.Close(); err != nil {
		return nil, fmt.Errorf("unexpected error closing udp connection while validating Goflow parameters: %w", err)
	}

	if c.Workers == 0 {
		c.Workers = 1
	}

	goflowInput := &GoflowInput{
		InputOperator: inputOperator,
		mode:          c.Mode,
		address:       c.Address,
		port:          int(c.Port),
		workers:       int(c.Workers),
	}
	return []operator.Operator{goflowInput}, nil
}

// GoflowInput is an operator that receives network traffic information
// from network devices.
type GoflowInput struct {
	helper.InputOperator
	wg     sync.WaitGroup
	cancel context.CancelFunc
	ctx    context.Context

	mode    string
	address string
	port    int
	workers int
}

// Start will start generating log entries.
func (n *GoflowInput) Start() error {
	n.ctx, n.cancel = context.WithCancel(context.Background())

	go func() {
		var goflowErr error
		var reuse = true

		backoff := backoff.Backoff{
			Min:    100 * time.Millisecond,
			Max:    3 * time.Second,
			Factor: 2,
			Jitter: false,
		}
		for {
			n.Infof("Starting Goflow on %s:%d in %s mode", n.address, n.port, n.mode)
			switch n.mode {
			case modeSflow:
				flow := &utils.StateSFlow{Transport: n, Logger: n}
				goflowErr = flow.FlowRoutine(n.workers, n.address, n.port, reuse)
			case modeNetflowV5:
				flow := &utils.StateNFLegacy{Transport: n, Logger: n}
				goflowErr = flow.FlowRoutine(n.workers, n.address, n.port, reuse)
			case modeNetflowV9:
				flow := &utils.StateNetFlow{Transport: n, Logger: n}
				goflowErr = flow.FlowRoutine(n.workers, n.address, n.port, reuse)
			}

			select {
			case <-n.ctx.Done():
				return
			default:
			}

			if goflowErr != nil {
				n.Errorf("Goflow quit with error", zap.Error(goflowErr))
			} else {
				n.Errorf("Goflow quit with unknown error")
			}

			time.Sleep(backoff.Duration())
			n.Infof("Restarting Goflow")
		}
	}()

	return nil
}

// Stop will stop generating logs.
func (n *GoflowInput) Stop() error {
	n.cancel()
	n.wg.Wait()
	return nil
}

// Publish writes entries and satisfies GoFlow's util.Transport interface
func (n *GoflowInput) Publish(messages []*flowmessage.FlowMessage) {
	n.wg.Add(1)
	defer n.wg.Done()

	for _, msg := range messages {
		m, t, err := Parse(*msg)
		if err != nil {
			n.Errorf("Failed to parse netflow message", zap.Error(err))
			continue
		}

		entry, err := n.NewEntry(m)
		if err != nil {
			n.Errorf("Failed to create new entry", zap.Error(err))
		}
		if !t.IsZero() {
			entry.Timestamp = t
		}
		n.Write(n.ctx, entry)
	}
}

// Printf is required by goflows logging interface
func (n *GoflowInput) Printf(format string, args ...interface{}) {
	n.Infof(format, args)
}