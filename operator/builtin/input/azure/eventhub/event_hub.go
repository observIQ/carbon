package eventhub

import (
	"context"
	"fmt"
	"sync"

	azhub "github.com/Azure/azure-event-hubs-go/v3"
	"github.com/observiq/stanza/operator"
	"github.com/observiq/stanza/operator/helper"
	"go.uber.org/zap"
)

const operatorName = "azure_event_hub_input"

func init() {
	operator.Register(operatorName, func() operator.Builder { return NewEventHubConfig("") })
}

// NewEventHubConfig creates a new Azure Event Hub input config with default values
func NewEventHubConfig(operatorID string) *EventHubInputConfig {
	return &EventHubInputConfig{
		InputConfig:   helper.NewInputConfig(operatorID, operatorName),
		PrefetchCount: 1000,
		StartAt:       "end",
	}
}

// EventHubInputConfig is the configuration of a Azure Event Hub input operator.
type EventHubInputConfig struct {
	helper.InputConfig `yaml:",inline"`

	// required
	Namespace        string `json:"namespace,omitempty"         yaml:"namespace,omitempty"`
	Name             string `json:"name,omitempty"              yaml:"name,omitempty"`
	Group            string `json:"group,omitempty"             yaml:"group,omitempty"`
	ConnectionString string `json:"connection_string,omitempty" yaml:"connection_string,omitempty"`

	// optional
	PrefetchCount uint32 `json:"prefetch_count,omitempty" yaml:"prefetch_count,omitempty"`
	StartAt       string `json:"start_at,omitempty"       yaml:"start_at,omitempty"`
}

// Build will build a Azure Event Hub input operator.
func (c *EventHubInputConfig) Build(buildContext operator.BuildContext) ([]operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(buildContext)
	if err != nil {
		return nil, err
	}

	if c.Namespace == "" {
		return nil, fmt.Errorf("missing required %s parameter 'namespace'", operatorName)
	}

	if c.Name == "" {
		return nil, fmt.Errorf("missing required %s parameter 'name'", operatorName)
	}

	if c.Group == "" {
		return nil, fmt.Errorf("missing required %s parameter 'group'", operatorName)
	}

	if c.ConnectionString == "" {
		return nil, fmt.Errorf("missing required %s parameter 'connection_string'", operatorName)
	}

	if c.PrefetchCount < 1 {
		return nil, fmt.Errorf("invalid value '%d' for %s parameter 'start_at'", c.PrefetchCount, operatorName)
	}

	var startAtEnd bool
	switch c.StartAt {
	case "beginning":
		startAtEnd = false
	case "end":
		startAtEnd = true
	default:
		return nil, fmt.Errorf("invalid value '%s' for %s parameter 'start_at'", c.StartAt, operatorName)
	}

	eventHubInput := &EventHubInput{
		InputOperator: inputOperator,
		namespace:     c.Namespace,
		name:          c.Name,
		group:         c.Group,
		connStr:       c.ConnectionString,
		prefetchCount: c.PrefetchCount,
		startAtEnd:    startAtEnd,
		persist: Persister{
			DB: helper.NewScopedDBPersister(buildContext.Database, c.ID()),
		},
	}
	return []operator.Operator{eventHubInput}, nil
}

// EventHubInput is an operator that reads input from Azure Event Hub.
type EventHubInput struct {
	helper.InputOperator
	cancel context.CancelFunc

	namespace     string
	name          string
	group         string
	connStr       string
	prefetchCount uint32
	startAtEnd    bool

	persist Persister
	hub     *azhub.Hub
	wg      sync.WaitGroup
}

// Start will start generating log entries.
func (e *EventHubInput) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	if err := e.persist.DB.Load(); err != nil {
		return err
	}

	hub, err := azhub.NewHubFromConnectionString(e.connStr, azhub.HubWithOffsetPersistence(&e.persist))
	if err != nil {
		return err
	}
	e.hub = hub

	runtimeInfo, err := hub.GetRuntimeInformation(ctx)
	if err != nil {
		return err
	}

	for _, partitionID := range runtimeInfo.PartitionIDs {
		if err := e.startConsumer(ctx, partitionID, hub); err != nil {
			return err
		}
		e.Infow(fmt.Sprintf("Successfully connected to Azure Event Hub '%s' partition_id '%s'", e.name, partitionID))
	}

	return nil
}

// Stop will stop generating logs.
func (e *EventHubInput) Stop() error {
	e.cancel()
	e.wg.Wait()
	if err := e.hub.Close(context.Background()); err != nil {
		return err
	}
	e.Infow(fmt.Sprintf("Closed all connections to Azure Event Hub '%s'", e.name))
	return e.persist.DB.Sync()
}

// poll starts polling an Azure Event Hub partition id for new events
func (e *EventHubInput) startConsumer(ctx context.Context, partitionID string, hub *azhub.Hub) error {
	// start at begining
	if !e.startAtEnd {
		_, err := hub.Receive(
			ctx, partitionID, e.handleEvent, azhub.ReceiveWithStartingOffset(""),
			azhub.ReceiveWithPrefetchCount(e.prefetchCount))
		return err
	}

	offsetStr := ""
	offset, err := e.persist.Read(e.namespace, e.name, e.group, partitionID)
	if err != nil {
		x := fmt.Sprintf("Error while reading offset for partition_id %s, starting at end", partitionID)
		e.Errorw(x, zap.Error(err))
	} else {
		offsetStr = offset.Offset
	}

	// start at end and no offset was found
	if offsetStr == "" {
		_, err := hub.Receive(
			ctx, partitionID, e.handleEvent, azhub.ReceiveWithLatestOffset(),
			azhub.ReceiveWithPrefetchCount(e.prefetchCount))
		return err
	}

	// start at end and offset exists
	_, err = hub.Receive(
		ctx, partitionID, e.handleEvent, azhub.ReceiveWithStartingOffset(offsetStr),
		azhub.ReceiveWithPrefetchCount(e.prefetchCount))
	return err
}

// handleEvents is the handler for hub.Receive.
func (e *EventHubInput) handleEvent(ctx context.Context, event *azhub.Event) error {
	e.wg.Add(1)

	// NewEntry wraps entry.New() and will attach labels and resources
	entry, err := e.NewEntry(nil)
	if err != nil {
		return err
	}

	entry, err = parseEvent(event, entry)
	if err != nil {
		return err
	}

	e.Write(ctx, entry)
	e.wg.Done()
	return nil
}
