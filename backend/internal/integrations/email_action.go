package integrations

import (
	"encoding/json"
	"fmt"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/rbac"
)

// SendEmailActionInput defines the parameters for the notifications.send_email action.
type SendEmailActionInput struct {
	ToRecipient    string `json:"to_recipient"`
	ToEmail        string `json:"to_email,omitempty"`
	Subject        string `json:"subject"`
	BodyText       string `json:"body_text"`
	BodyHTML       string `json:"body_html,omitempty"`
	FromEmail      string `json:"from_email,omitempty"`
	CorrelationID  string `json:"correlation_id,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// SendEmailAction integrates Email dispatching into the LogisticsHQ Action System.
type SendEmailAction struct {
	service GatewayService
}

// NewSendEmailAction creates an action instance for dispatching Email via AWS SES.
func NewSendEmailAction(service GatewayService) actions.Action {
	return &SendEmailAction{service: service}
}

func (a *SendEmailAction) Name() string {
	return "notifications.send_email"
}

func (a *SendEmailAction) Module() string {
	return "notifications"
}

func (a *SendEmailAction) Description() string {
	return "Dispatch an outbound Email notification to a verified recipient via external AWS SES gateway."
}

func (a *SendEmailAction) Category() actions.ActionCategory {
	return actions.ActionCategoryWrite
}

func (a *SendEmailAction) InputSchema() interface{} {
	return &SendEmailActionInput{}
}

func (a *SendEmailAction) RequiresConfirmation() bool {
	return false
}

func (a *SendEmailAction) RequiredPermission() (string, string) {
	return rbac.ResourceOutreach, rbac.ActionCreate
}

func (a *SendEmailAction) Execute(ctx *actions.ActionContext, input []byte) (*actions.ActionResult, error) {
	var in SendEmailActionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: fmt.Sprintf("invalid Email action payload: %v", err),
			},
		}, nil
	}

	correlationID := in.CorrelationID
	var orgID int64
	idempotencyKey := in.IdempotencyKey
	if ctx != nil {
		orgID = ctx.OrganizationID
		if correlationID == "" {
			correlationID = ctx.RequestID
		}
		if idempotencyKey == "" && ctx.IdempotencyKey != nil {
			idempotencyKey = *ctx.IdempotencyKey
		}
	}

	to := in.ToRecipient
	if to == "" {
		to = in.ToEmail
	}

	resp, err := a.service.SendEmail(ctx.Context, orgID, EmailRequest{
		ToRecipient:    to,
		Subject:        in.Subject,
		BodyText:       in.BodyText,
		BodyHTML:       in.BodyHTML,
		FromEmail:      in.FromEmail,
		CorrelationID:  correlationID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "ExecutionFailed",
				Message: err.Error(),
			},
		}, nil
	}

	respBytes, _ := json.Marshal(resp)
	return &actions.ActionResult{
		Success: true,
		Action:  a.Name(),
		Data:    respBytes,
	}, nil
}
