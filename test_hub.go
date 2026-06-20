package main
import (
	"encoding/json"
	"fmt"
	"time"
	"meownet-core/models"
)
func main() {
	instance := models.RoomInstance{
		Id: 3,
		RoomId: 1,
		SubRoomId: 1,
		Location: "c79709d8-a31b-48aa-9eb8-cc31ba9505e8",
		PhotonRegionId: "us",
		PhotonRoomId: "2edbdcdb-58d3-4ac9-9283-7519f1d2ec1b",
		Name: "Orientation",
		MaxCapacity: 40,
		CreatedAt: time.Now(),
	}
	b, _ := json.MarshalIndent(instance, "", "  ")
	fmt.Println(string(b))
}
