package cmd

type Command struct {
	Action  Action `json:"action"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

type Action string

const (
	ActionCreate  Action = "create"
	ActionRead    Action = "read"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionConfirm Action = "yes"
	ActionCancel  Action = "no"
	ActionUnknown Action = ""
	NameAll              = "A-L-L"
)
