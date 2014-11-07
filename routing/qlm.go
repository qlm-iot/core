package routing

import (
	"github.com/qlm-iot/qlm/mi"
)

func Process(msg []byte, db Datastore) {
	envelope, err := mi.Unmarshal(msg)
	if err == nil {
		if len(envelope.Cancel.RequestIds) > 0 {
			for _, rId := range envelope.Cancel.RequestIds {
				db.Cancel(rId.Text)
			}
		}

		var read = &envelope.Read
		if read != nil {
		}

		var write = &envelope.Write
		if write != nil {
		}

	}
}
