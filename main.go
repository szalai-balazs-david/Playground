// OPC UA client working against Unified Automation ANSI C Demo OPC UA server: https://www.unified-automation.com/downloads/opc-ua-servers.html
// The app subscribes to a set of Node IDs and logs their value to a SQLite database

package main

import (
	"context"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"gorm.io/gorm"
)

type OpcData struct {
	gorm.Model
	Timestamp time.Time
	Double    float64
	Float     float32
	Int32     int32
	Running   bool
	Speed     uint32
}

func main() {
	server := "opc.tcp://localhost:48020"
	nodes := []string{
		"ns=4;s=Demo.Dynamic.Scalar.Double",
		"ns=4;s=Demo.Dynamic.Scalar.Float",
		"ns=4;s=Demo.Dynamic.Scalar.Int32",
		"ns=4;s=Demo.SimulationActive",
		"ns=4;s=Demo.SimulationSpeed",
	}

	db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&OpcData{})
	ctx := context.Background()

	c, err := opcua.NewClient(server)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close(ctx)

	notifyCh := make(chan *opcua.PublishNotificationData)

	sub, err := c.Subscribe(ctx, &opcua.SubscriptionParameters{
		Interval: 50 * time.Millisecond, // Oddly, this works. I thought OPC UA is 10Hz max.
	}, notifyCh)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Cancel(ctx)
	log.Printf("Created subscription with id %v", sub.SubscriptionID)

	items := []*ua.MonitoredItemCreateRequest{}
	for i, nodeId := range nodes {
		id, err := ua.ParseNodeID(nodeId)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, opcua.NewMonitoredItemCreateRequestWithDefaults(id, ua.AttributeIDValue, uint32(i)))
	}
	res, err := sub.Monitor(ctx, ua.TimestampsToReturnBoth, items...)
	if err != nil || res.Results[0].StatusCode != ua.StatusOK {
		log.Fatal(err)
	}

	//OPC UA subscriptions works with change events, we only get updates from things that changed
	//As Running and Speed are changing infrequently, we have to manage a cache where we store the "latest" value of each item
	//Downside is that now we have to manage 2 copies of the data: The new database entry and the cache
	//Using only the cache and not creating an OpcData item in the subscription callback results in DB ID conflict: GORM is not auto-incrementing the unique ID field unless we create a new struct
	var cache OpcData
	for {
		select {
		case <-ctx.Done():
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
				//    a) the timestamp won't be correct: We assign the 1st timestamp, but due to loop-processing, the values from the "latest" event will be present
				//    b) We lose data: We only create 1 DB entry even though we received 1+ change sets
				//I've set the subscription interval to 50ms, which is sufficient for the server's 65-70ms change rate, but architecturally this is not sound.
				data := OpcData{
					Timestamp: x.MonitoredItems[0].Value.ServerTimestamp,
					Running:   cache.Running,
					Speed:     cache.Speed,
					Double:    cache.Double,
					Float:     cache.Float,
					Int32:     cache.Int32,
				}
				for _, item := range x.MonitoredItems {
					//This is a naive approach that works due to each node ID that's been subscribed to having unique data types.
					//More realistically I should map by ClientHandle, which was earlier defined when adding the monitored items to the subscription.
					switch itemValue := item.Value.Value.Value().(type) {
					case float64:
						data.Double = itemValue
						cache.Double = itemValue
					case float32:
						data.Float = itemValue
						cache.Float = itemValue
					case int32:
						data.Int32 = itemValue
						cache.Int32 = itemValue
					case uint32:
						data.Speed = itemValue
						cache.Speed = itemValue
					case bool:
						data.Running = itemValue
						cache.Running = itemValue
					}
				}
				log.Printf("%v", len(x.MonitoredItems))
				db.Create(&data)

			default:
				log.Printf("Unknown data type: %T", res.Value)
			}
		}
	}
}
