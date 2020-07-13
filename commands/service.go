package commands

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kardianos/service"
	"github.com/observiq/carbon/agent"
	"go.uber.org/zap"
)

// AgentService is a service that runs the carbon agent.
type AgentService struct {
	cancel context.CancelFunc
	agent  *agent.LogAgent
}

// Start will start the carbon agent.
func (a *AgentService) Start(s service.Service) error {
	a.agent.Info("Starting carbon agent")
	if err := a.agent.Start(); err != nil {
		a.agent.Errorw("Failed to start carbon agent", zap.Any("error", err))
		a.cancel()
	}
	return nil
}

// Stop will stop the carbon agent.
func (a *AgentService) Stop(s service.Service) error {
	a.agent.Info("Stopping carbon agent")
	a.agent.Stop()
	a.cancel()
	return nil
}

// newAgentService creates a new agent service with the provided agent.
func newAgentService(ctx context.Context, agent *agent.LogAgent, cancel context.CancelFunc) (service.Service, error) {
	agentService := &AgentService{cancel, agent}
	config := &service.Config{
		Name:        "carbon",
		DisplayName: "Carbon Log Agent",
		Description: "Monitors and processes log entries",
		Option: service.KeyValue{
			"RunWait": func() {
				var sigChan = make(chan os.Signal, 3)
				signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
				select {
				case <-sigChan:
				case <-ctx.Done():
				}
			},
		},
	}

	service, err := service.New(agentService, config)
	if err != nil {
		return nil, err
	}

	return service, nil
}
