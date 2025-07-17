package queue

type QueueConfig struct {
	Exchange   string
	Queue      string
	RoutingKey string
}

var List = map[string]QueueConfig{
	"logs": {
		Exchange:   "amq.direct",
		Queue:      "sso:logs",
		RoutingKey: "JIUzUxMiIs",
	},
	"usrRegSignal": {
		Exchange:   "amq.direct",
		Queue:      "register:user-registration-signal",
		RoutingKey: "cCI6IkpXVC",
	},
}
