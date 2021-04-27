package azure

import (
	"context"
	"fmt"
	"sync"

	azhub "github.com/Azure/azure-event-hubs-go/v3"
	"github.com/observiq/stanza/operator/helper"
	"go.uber.org/zap"
)

// Eventhub provides methods for reading events from Azure Event Hub.
type EventHub struct {
	Namespace     string
	Name          string
	Group         string
	ConnStr       string
	PrefetchCount uint32
	StartAtEnd    bool
	Persist       *Persister
	WG            sync.WaitGroup
	Handler       func(context.Context, *azhub.Event) error

	hub *azhub.Hub
	helper.InputOperator
}

// StartConsumers starts an Azure Event Hub handler for each partition_id.
func (e *EventHub) StartConsumers(ctx context.Context) error {
	if err := e.Persist.DB.Load(); err != nil {
		return err
	}

	if err := e.Connect(); err != nil {
		return err
	}

	runtimeInfo, err := e.hub.GetRuntimeInformation(ctx)
	if err != nil {
		return err
	}

	for _, partitionID := range runtimeInfo.PartitionIDs {
		if err := e.startConsumer(ctx, partitionID, e.hub); err != nil {
			return err
		}
		e.Debugw(fmt.Sprintf("Successfully connected to Azure Event Hub '%s' partition_id '%s'", e.Name, partitionID))
	}

	return nil
}

// StopConsumers closes connections to Azure Event Hub.
func (e *EventHub) StopConsumers() error {
	if e.hub == nil {
		return nil
	}
	e.WG.Wait()
	if err := e.hub.Close(context.Background()); err != nil {
		return err
	}
	e.Debugw(fmt.Sprintf("Closed all connections to Azure Event Hub '%s'", e.Name))
	if err := e.Persist.DB.Sync(); err != nil {
		return err
	}
	return nil
}

// startConsumer starts polling an Azure Event Hub partition id for new events
func (e *EventHub) startConsumer(ctx context.Context, partitionID string, hub *azhub.Hub) error {
	// start at begining
	if !e.StartAtEnd {
		_, err := hub.Receive(
			ctx, partitionID, e.Handler, azhub.ReceiveWithStartingOffset(""),
			azhub.ReceiveWithPrefetchCount(e.PrefetchCount))
		return err
	}

	offset, err := e.Persist.Read(e.Namespace, e.Name, e.Group, partitionID)
	if err != nil {
		x := fmt.Sprintf("Error while reading offset for partition_id %s", partitionID)
		e.Debugw(x, zap.Error(err))
	}

	// start at end and no offset was found
	if offset.Offset == "" {
		e.Debugw("No offset found, starting from end")
		_, err := hub.Receive(
			ctx, partitionID, e.Handler, azhub.ReceiveWithLatestOffset(),
			azhub.ReceiveWithPrefetchCount(e.PrefetchCount))
		return err
	}

	// start at end and offset exists
	e.Debugw(fmt.Sprintf("Starting with offset '%s'", offset.Offset))
	_, err = hub.Receive(
		ctx, partitionID, e.Handler, azhub.ReceiveWithStartingOffset(offset.Offset),
		azhub.ReceiveWithPrefetchCount(e.PrefetchCount))
	return err
}

// Connect initializes the connection to Azure Event Hub ensures the input parameters are valid
func (e *EventHub) Connect() error {
	var err error
	e.hub, err = azhub.NewHubFromConnectionString(e.ConnStr, azhub.HubWithOffsetPersistence(e.Persist))
	return err
}
