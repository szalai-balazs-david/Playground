package main

import (
	"context"
	"log"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

const opcUaServer = "opc.tcp://localhost:48020"
const (
	speedHandle   = iota
	runningHandle = iota
	doubleHandle  = iota
	floatHandle   = iota
	int32Handle   = iota
)

func RunOpcMonitoring(ctx context.Context, ds *DataStore) {
	nodes := map[int32]string{
		speedHandle:   "ns=4;s=Demo.SimulationSpeed",
		runningHandle: "ns=4;s=Demo.SimulationActive",
		doubleHandle:  "ns=4;s=Demo.Dynamic.Scalar.Double",
		floatHandle:   "ns=4;s=Demo.Dynamic.Scalar.Float",
		int32Handle:   "ns=4;s=Demo.Dynamic.Scalar.Int32",
	}
	c, err := opcua.NewClient(opcUaServer)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connecting...")
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close(ctx)

	notifyCh := make(chan *opcua.PublishNotificationData)

	log.Print("Creating subscription...")
	sub, err := c.Subscribe(ctx, &opcua.SubscriptionParameters{
		Interval: 50 * time.Millisecond, // Oddly, this works. I thought OPC UA is 10Hz max.
	}, notifyCh)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Cancel(ctx)
	log.Printf("Created subscription with id %v", sub.SubscriptionID)

	items := []*ua.MonitoredItemCreateRequest{}
	for clientHandle, nodeId := range nodes {
		id, err := ua.ParseNodeID(nodeId)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, opcua.NewMonitoredItemCreateRequestWithDefaults(id, ua.AttributeIDValue, uint32(clientHandle)))
	}

	log.Print("Starting monitoring...")
	res, err := sub.Monitor(ctx, ua.TimestampsToReturnBoth, items...)
	if err != nil || res.Results[0].StatusCode != ua.StatusOK {
		log.Fatal(err)
	}

	//OPC UA subscriptions works with change events, we only get updates from things that changed
	//As "Running" and "Speed" are changing infrequently, we have to manage a cache where we store the "latest" value of each item
	//Downside is that now we have to manage 2 copies of the data: The new database entry and the cache
	//Using only the cache and not creating an OpcData item in the subscription callback results in DB ID conflict: GORM is not auto-incrementing the unique ID field unless we create a new struct
	var cache OpcData
	for {
		select {
		case <-ctx.Done():
			log.Print("Shutting down OPC routine")
			c.Close(ctx)
			return
		case res := <-notifyCh:
			if res.Error != nil {
				log.Print(res.Error)
				continue
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				//Note: There's an issue here. The server might send multiple change events per single node. It happens if the subscription interval is set to higher than the rate of change of the values.
				//The problem is that in that case
				//    a) the timestamp won't be correct: We assign the 1st timestamp, but due to loop-processing, the values from the "latest" event will be stored
				//    b) We lose data: We only create 1 DB entry even though we received 1+ change sets
				//I've set the subscription interval to 50ms, which is sufficient for the server's 65-70ms change rate, but architecturally this is not sound.
				data := OpcData{
					Timestamp: x.MonitoredItems[0].Value.ServerTimestamp.UTC(),
					Running:   cache.Running,
					Speed:     cache.Speed,
					Double:    cache.Double,
					Float:     cache.Float,
					Int32:     cache.Int32,
				}
				for _, item := range x.MonitoredItems {
					//This is tedious to maintain. Maybe there's a better way?
					switch item.ClientHandle {
					case doubleHandle:
						val := item.Value.Value.Value().(float64)
						data.Double = val
						cache.Double = val
					case floatHandle:
						val := item.Value.Value.Value().(float32)
						data.Float = val
						cache.Float = val
					case int32Handle:
						val := item.Value.Value.Value().(int32)
						data.Int32 = val
						cache.Int32 = val
					case speedHandle:
						val := item.Value.Value.Value().(uint32)
						data.Speed = val
						cache.Speed = val
					case runningHandle:
						val := item.Value.Value.Value().(bool)
						data.Running = val
						cache.Running = val
					}
				}
				ds.Add(&data)

			default:
				log.Printf("Unknown data type: %T", res.Value)
			}
		}
	}
}
