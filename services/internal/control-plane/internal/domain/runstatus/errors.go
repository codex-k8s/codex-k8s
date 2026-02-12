package runstatus

import "errors"

var (
	errRunNotFound              = errors.New("run not found")
	errRunPayloadEmpty          = errors.New("run payload is empty")
	errRunPayloadDecode         = errors.New("run payload decode failed")
	errRunIssueNumberMissing    = errors.New("run payload issue number is required")
	errRunRepoNameMissing       = errors.New("run payload repository full_name is required")
	errRunRepoTokenMissing      = errors.New("repository token is missing")
	errRunRepoTokenDecrypt      = errors.New("repository token decrypt failed")
	errRunRepoBindingRequired   = errors.New("run payload project.repository_id is required")
	errRunStatusCommentNotFound = errors.New("run status comment not found")
	errRunNamespaceMissing      = errors.New("run namespace is missing")
)
