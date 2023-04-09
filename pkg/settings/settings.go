package settings

var (
	DataPathPrefix    = "./data/"
	RaftRPCListenPort = ":9000"
	// actual replica count will be determined on the fly with a adaptive
	// algorithm implemented
	DefaultReplicaCount = 3
)
