package runstatus

import "errors"

var (
	errRunNotFound              = errors.New("run not found")
	errRunPayloadEmpty          = errors.New("run payload is empty")
	errRunPayloadDecode         = errors.New("run payload decode failed")
	errRunCommentTargetMissing  = errors.New("run payload comment target is required")
	errRunRepoNameMissing       = errors.New("run payload repository full_name is required")
	errRunBotTokenMissing       = errors.New("bot token is missing")
	errRunBotTokenDecrypt       = errors.New("bot token decrypt failed")
	errRunStatusCommentNotFound = errors.New("run status comment not found")
	errRunNamespaceMissing      = errors.New("run namespace is missing")
)
