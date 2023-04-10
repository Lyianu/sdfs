package settings

const (
	// path for client to download
	URLDownload = "/api/download"
	// path for client to request upload
	URLUpload = "/api/upload"
	// path for client to request delete
	URLDelete = "/api/delete"
	// path for sdfs master to request download
	URLSDFSDownload = "/api/sdfs/download"
	// path for sdfs master to request delete
	URLSDFSDelete = "/api/sdfs/delete"
	// path for sdfs master to request upload
	URLSDFSUpload = "/api/sdfs/upload"
	// path for sdfs node to send heartbeat to
	URLSDFSHeartbeat = "/api/sdfs/heartbeat"

	// not implemented
	URLSDFSWrite = "/api/sdfs/write"
	URLWrite     = "/api/write"

	URLDebugPrintHashstore = "/api/debug/prinths"
	URLDebugPrintSDFS      = "/api/debug/printfs"

	// path for sdfs master to request replica creation and node to callback
	URLSDFSReplicaRequest  = "/api/sdfs/replica/request"
	URLSDFSReplicaCallback = "/api/sdfs/replica/callback"

	URLUploadCallback = "/api/callback/upload"

	// the scheme that sdfs master uses, http or https
	URLSDFSScheme = "http://"
)
