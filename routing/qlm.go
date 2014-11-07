package routing

import (
	"github.com/qlm-iot/qlm"
)

func Process(mchan chan []byte, db Datastore) {
	for {
		var msg []byte
		select {
		case msg = <-mchan:
			answer, _ := qlm.Unmarshal(msg)
			for _, data := range answer.Objects {
				for _, info := range data.InfoItems {
					fmt.Println(info.Name)
				}
			}
			fmt.Println(answer.Version)
		}
	}
}
