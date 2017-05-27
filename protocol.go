package dryred

const FWConnectRequestName = "connect"
const FWListenRequestName = "listen"

type FWConnectRequest_V1 struct {
	BackClientName string
	BackConnectionAddress string
}

type FWConnectReply_V1 struct {
	Status bool
	ErrorMessage string
}

type FWListenRequest_V1 struct {
	Name string
}

type FWListenReply_V1 struct {
	Status bool
	ErrorMessage string
}
