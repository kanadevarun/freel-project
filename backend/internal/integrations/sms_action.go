package integrations

import (
	"encoding/json"
	"fmt"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/rbac"
)

// SendSMSActionInput defines the parameters for the notifications.send_sms action.
type SendSMSActionInput struct {
	RecipientPhoneNumber string `json:"recipient_phone_number"`
	Message              string `json:"message"`
	SenderID             string `json:"sender_id,omitempty"`
	CorrelationID        string `json:"correlation_id,omitempty"`
	IdempotencyKey       string `json:"idempotency_key,omitempty"`
}

// SendSMSAction integrates SMS dispatching into the LogisticsHQ Action System.
type SendSMSAction struct {
	service GatewayService
}

// NewSendSMSAction creates an action instance for dispatching SMS.
func NewSendSMSAction(service GatewayService) actions.Action {
	return &SendSMSAction{service: service}
}

func (a *SendSMSAction) Name() string {
	return "notifications.send_sms"
}

func (a *SendSMSAction) Module() string {
	return "notifications"
}

func (a *SendSMSAction) Description() string {
	return "Dispatch an outbound SMS notification to a verified recipient via external Twilio gateway."
}

func (a *SendSMSAction) Category() actions.ActionCategory {
	return actions.ActionCategoryWrite
}

func (a *SendSMSAction) InputSchema() interface{} {
	return &SendSMSActionInput{}
}

func (a *SendSMSAction) RequiresConfirmation() bool {
	return false
}

func (a *SendSMSAction) RequiredPermission() (string, string) {
	return rbac.ResourceOutreach, rbac.ActionCreate
}

func (a *SendSMSAction) Execute(ctx *actions.ActionContext, input []byte) (*actions.ActionResult, error) {
	var in SendSMSActionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: fmt.Sprintf("invalid SMS action payload: %v", err),
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

	resp, err := a.service.SendSMS(ctx.Context, orgID, SMSRequest{
		RecipientPhoneNumber: in.RecipientPhoneNumber,
		Body:                 in.Message,
		SenderID:             in.SenderID,
		CorrelationID:        correlationID,
		IdempotencyKey:       idempotencyKey,
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

	return &actions.ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "sms",
		ResourceID:   resp.MessageID,
		Summary:      fmt.Sprintf("Outbound SMS accepted by provider %s (Status: %s)", resp.Provider, resp.Status),
		Data:         resp,
	}, nil
}
