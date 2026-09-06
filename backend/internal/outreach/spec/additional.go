package spec

import "time"

// Audience Models

type AudienceLead struct {
	ID          int32     `json:"id" db:"id"`
	CompanyName string    `json:"company_name" db:"company_name"`
	ContactName *string   `json:"contact_name" db:"contact_name"`
	Email       *string   `json:"email" db:"email"`
	Status      string    `json:"status" db:"status"`
	AddedAt     time.Time `json:"added_at" db:"added_at"`
}

type GetCampaignAudienceRequest struct {
	CampaignID int32 `json:"-"`
	OrgID      int32 `json:"-"`
}

type AddCampaignAudienceRequest struct {
	CampaignID int32   `json:"-"`
	OrgID      int32   `json:"-"`
	LeadIDs    []int32 `json:"lead_ids"`
}

type RemoveCampaignAudienceRequest struct {
	CampaignID int32 `json:"-"`
	LeadID     int32 `json:"-"`
	OrgID      int32 `json:"-"`
}

type GetCampaignAudienceResponse struct {
	Leads []*AudienceLead `json:"leads"`
}

// Sequence Step Requests

type AddSequenceStepRequest struct {
	CampaignID int32  `json:"-"`
	OrgID      int32  `json:"-"`
	Name       string `json:"name"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
	DelayDays  int    `json:"delay_days"`
	Channel    string `json:"channel"`
}

type UpdateSequenceStepRequest struct {
	CampaignID int32  `json:"-"`
	StepID     int32  `json:"-"`
	OrgID      int32  `json:"-"`
	Name       string `json:"name"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
	DelayDays  int    `json:"delay_days"`
	Channel    string `json:"channel"`
}

type DeleteSequenceStepRequest struct {
	CampaignID int32 `json:"-"`
	StepID     int32 `json:"-"`
	OrgID      int32 `json:"-"`
}

type ReorderSequenceRequest struct {
	CampaignID int32   `json:"-"`
	OrgID      int32   `json:"-"`
	StepIDs    []int32 `json:"step_ids"`
}

// Analytics and Insights Responses

type OutreachAnalyticsResponse struct {
	TotalCampaigns      int      `json:"total_campaigns"`
	ActiveCampaigns     int      `json:"active_campaigns"`
	PausedCampaigns     int      `json:"paused_campaigns"`
	CompletedCampaigns  int      `json:"completed_campaigns"`
	TotalRecipients     int      `json:"total_recipients"`
	ActiveRecipients    int      `json:"active_recipients"`
	CompletedRecipients int      `json:"completed_recipients"`
	AwaitingRecipients  int      `json:"awaiting_recipients"`
	LeadsGenerated      int      `json:"leads_generated"`
	ConversionRate      float64  `json:"conversion_rate"`
	EmailsSent          *int     `json:"emails_sent"`
	EmailsDelivered     *int     `json:"emails_delivered"`
	EmailsOpened        *int     `json:"emails_opened"`
	LinksClicked        *int     `json:"links_clicked"`
	RepliesReceived     *int     `json:"replies_received"`
	Bounces             *int     `json:"bounces"`
	Unsubscribes        *int     `json:"unsubscribes"`
}

type FunnelStage struct {
	Stage string `json:"stage"`
	Count int    `json:"count"`
}

type ConversionFunnelResponse struct {
	Stages []FunnelStage `json:"stages"`
}

type CampaignInsight struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // 'WARNING', 'INFO', 'CRITICAL'
	Action      string `json:"action"`
}

type CampaignInsightsResponse struct {
	Insights []CampaignInsight `json:"insights"`
}

type GeneratedLead struct {
	ID             int32     `json:"id" db:"id"`
	CompanyName    string    `json:"company_name" db:"company_name"`
	ContactName    *string   `json:"contact_name" db:"contact_name"`
	Email          *string   `json:"email" db:"email"`
	Status         string    `json:"status" db:"status"`
	ConvertedAt    time.Time `json:"converted_at" db:"converted_at"`
	AssignedToName *string   `json:"assigned_to_name" db:"assigned_to_name"`
}

type CampaignLeadsResponse struct {
	Leads []*GeneratedLead `json:"leads"`
}

// Outreach Dashboard Overview Models
type OutreachDashboardResponse struct {
	// Top Stats
	ActiveOutreach       int `json:"active_outreach"`
	DueToday             int `json:"due_today"`
	Overdue              int `json:"overdue"`
	Completed            int `json:"completed"`
	OpportunitiesCreated int `json:"opportunities_created"`
	ActiveCampaigns      int `json:"active_campaigns"`
	ActiveProspects      int `json:"active_prospects"`
	EngagedProspects     int `json:"engaged_prospects"`
	RepliesCount         int `json:"replies_count"`

	// Outreach by Type (counts)
	TypeEmail    int `json:"type_email"`
	TypeCall     int `json:"type_call"`
	TypeFollowup int `json:"type_followup"`
	TypeMeeting  int `json:"type_meeting"`
	TypeOther    int `json:"type_other"`

	// Outreach by Status (counts)
	StatusPending    int `json:"status_pending"`
	StatusInProgress int `json:"status_in_progress"`
	StatusCompleted  int `json:"status_completed"`
	StatusOverdue    int `json:"status_overdue"`

	// Lists
	UpcomingFollowups []*OutreachActivityDetail `json:"upcoming_followups"`
	RecentActivities  []*OutreachActivityDetail `json:"recent_activities"`
	OverdueItems      []*OutreachActivityDetail `json:"overdue_items"`
}

type OutreachActivityDetail struct {
	ID           int64      `json:"id" db:"id"`
	OrgID        int64      `json:"org_id" db:"org_id"`
	LeadID       *int64     `json:"lead_id" db:"lead_id"`
	CustomerID   *int64     `json:"customer_id" db:"customer_id"`
	ActivityType string     `json:"activity_type" db:"activity_type"`
	Subject      string     `json:"subject" db:"subject"`
	Description  *string    `json:"description" db:"description"`
	Status       string     `json:"status" db:"status"`
	Priority     string     `json:"priority" db:"priority"`
	ScheduledAt  *time.Time `json:"scheduled_at" db:"scheduled_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy    *int64     `json:"created_by" db:"created_by"`

	// Joined/computed fields
	LeadCompanyName     *string `json:"lead_company_name" db:"lead_company_name"`
	LeadContactName     *string `json:"lead_contact_name" db:"lead_contact_name"`
	LeadEmail           *string `json:"lead_email" db:"lead_email"`
	CustomerCompanyName *string `json:"customer_company_name" db:"customer_company_name"`
	CustomerContactName *string `json:"customer_contact_name" db:"customer_contact_name"`
	CreatorName         *string `json:"creator_name" db:"creator_name"`
}

type CreateActivityRequest struct {
	OrgID        int32      `json:"-"`
	LeadID       *int64     `json:"lead_id"`
	CustomerID   *int64     `json:"customer_id"`
	ActivityType string     `json:"activity_type"`
	Subject      string     `json:"subject"`
	Description  *string    `json:"description"`
	Status       string     `json:"status"`
	Priority     string     `json:"priority"`
	ScheduledAt  *time.Time `json:"scheduled_at"`
	CreatedBy    *int64     `json:"-"`
}

type UpdateActivityRequest struct {
	ID           int64      `json:"-"`
	OrgID        int32      `json:"-"`
	LeadID       *int64     `json:"lead_id"`
	CustomerID   *int64     `json:"customer_id"`
	ActivityType string     `json:"activity_type"`
	Subject      string     `json:"subject"`
	Description  *string    `json:"description"`
	Status       string     `json:"status"`
	Priority     string     `json:"priority"`
	ScheduledAt  *time.Time `json:"scheduled_at"`
}

type GetActivityRequest struct {
	ID    int64 `json:"-"`
	OrgID int32 `json:"-"`
}

type CampaignRecipient struct {
	LeadID           int64      `json:"lead_id" db:"lead_id"`
	CompanyName      string     `json:"company_name" db:"company_name"`
	ContactName      *string    `json:"contact_name" db:"contact_name"`
	Email            *string    `json:"email" db:"email"`
	Phone            *string    `json:"phone" db:"phone"`
	LeadStatus       string     `json:"lead_status" db:"lead_status"`
	EngagementStatus string     `json:"engagement_status"`
	CampaignID       *int64     `json:"campaign_id" db:"campaign_id"`
	CampaignName     *string    `json:"campaign_name" db:"campaign_name"`
	Status           string     `json:"status" db:"status"`
	CurrentStep      int        `json:"current_step" db:"current_step"`
	EmailsSent       int        `json:"emails_sent"`
	LastActivityAt   *time.Time `json:"last_activity_at" db:"last_activity_at"`
	LastActivityDesc *string    `json:"last_activity_desc"`
	NextScheduledAt  *time.Time `json:"next_scheduled_at" db:"next_scheduled_at"`
	OwnerName        *string    `json:"owner_name" db:"owner_name"`
}

type CampaignActivityEvent struct {
	ID              int64      `json:"id" db:"id"`
	ActivityType    string     `json:"activity_type" db:"activity_type"`
	Subject         string     `json:"subject" db:"subject"`
	Description     *string    `json:"description" db:"description"`
	Status          string     `json:"status" db:"status"`
	ScheduledAt     *time.Time `json:"scheduled_at" db:"scheduled_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	LeadCompanyName *string    `json:"lead_company_name" db:"lead_company_name"`
	LeadContactName *string    `json:"lead_contact_name" db:"lead_contact_name"`
	CreatorName     *string    `json:"creator_name" db:"creator_name"`
}

type ProspectEngagementResponse struct {
	LeadID           int64                    `json:"lead_id"`
	CompanyName      string                   `json:"company_name"`
	ContactName      *string                  `json:"contact_name"`
	Email            *string                  `json:"email"`
	LeadStatus       string                   `json:"lead_status"`
	EngagementStatus string                   `json:"engagement_status"`
	CampaignName     *string                  `json:"campaign_name"`
	EmailsSent       int                      `json:"emails_sent"`
	LastActivityAt   *time.Time               `json:"last_activity_at"`
	Timeline         []*CampaignActivityEvent `json:"timeline"`
}

type ProspectDetailResponse struct {
	Prospect      *CampaignRecipient       `json:"prospect"`
	Campaigns     []*Campaign              `json:"campaigns"`
	SequenceSteps []*Sequence              `json:"sequence_steps"`
	Activities    []*CampaignActivityEvent `json:"activities"`
	FollowUps     []*OutreachActivityDetail `json:"follow_ups"`
}

type EnrollProspectRequest struct {
	CampaignID int32 `json:"campaign_id"`
	LeadID     int64 `json:"lead_id"`
}

type UpdateProspectRequest struct {
	LeadID      int64  `json:"-"`
	OrgID       int32  `json:"-"`
	CampaignID  int32  `json:"campaign_id"`
	Status      string `json:"status"`
	CurrentStep int    `json:"current_step"`
}

type RescheduleFollowUpRequest struct {
	ID          int64      `json:"-"`
	OrgID       int32      `json:"-"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}


