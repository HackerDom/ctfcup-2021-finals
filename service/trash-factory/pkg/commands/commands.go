package commands

const (
	ContainerCreate  = '\x01'
	ContainerList    = '\x02'
	GetContainerInfo = '\x03'
	PutItem          = '\x04'
	GetItem          = '\x05'
	CreateUser       = '\x06'
	GetUser          = '\x07'
	ListUsers        = '\x08'
	GetStatistic     = '\x09'
	GetAdminToken    = '\x10'
)

const (
	StatusOK                 = '\x00'
	StatusIncorrectSignature = '\x01'
	StatusFunctionNotFound   = '\x02'
	StatusCommandExecError   = '\x03'
	StatusCommandNotFound    = '\x04'
	StatusInternalError      = '\xff'
)
