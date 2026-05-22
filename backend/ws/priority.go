package ws

const (
	PriorityNone   = 0
	PriorityAngel  = 1
	PriorityAlice  = 2
)

func BrokerPriority(name string) int {
	switch name {
	case "aliceblue":
		return PriorityAlice
	case "angel":
		return PriorityAngel
	default:
		return PriorityNone
	}
}
