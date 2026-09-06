package outreach

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/freel/backend/internal/outreach/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	CreateCampaign(ctx context.Context, c *spec.Campaign) error
	GetCampaignByID(ctx context.Context, orgID, id int32) (*spec.Campaign, error)
	ListCampaigns(ctx context.Context, orgID int32, limit, offset int) ([]*spec.Campaign, int, error)
	UpdateCampaign(ctx context.Context, c *spec.Campaign) error
	DeleteCampaign(ctx context.Context, orgID, id int32) error
	AddSequence(ctx context.Context, s *spec.Sequence) error

	// Sequence Step Methods
	GetCampaignSequence(ctx context.Context, orgID, campaignID int32) ([]*spec.Sequence, error)
	AddSequenceStep(ctx context.Context, s *spec.Sequence) error
	UpdateSequenceStep(ctx context.Context, s *spec.Sequence) error
	DeleteSequenceStep(ctx context.Context, campaignID, stepID int32) error
	ReorderSequence(ctx context.Context, campaignID int32, stepIDs []int32) error

	// Audience Methods
	GetCampaignAudience(ctx context.Context, orgID, campaignID int32) ([]*spec.AudienceLead, error)
	AddCampaignAudience(ctx context.Context, campaignID int32, leadIDs []int32) error
	RemoveCampaignAudience(ctx context.Context, campaignID, leadID int32) error

	// Analytics Methods
	GetOutreachAnalytics(ctx context.Context, orgID int32) (*spec.OutreachDashboardResponse, error)
	GetCampaignAnalytics(ctx context.Context, orgID, campaignID int32) (*spec.OutreachAnalyticsResponse, error)
	GetCampaignLeads(ctx context.Context, orgID, campaignID int32) ([]*spec.GeneratedLead, error)
	GetCampaignInsights(ctx context.Context, orgID, campaignID int32) ([]spec.CampaignInsight, error)
	GetConversionFunnel(ctx context.Context, orgID int32) (*spec.ConversionFunnelResponse, error)

	// Activity CRUD Methods
	CreateActivity(ctx context.Context, req *spec.CreateActivityRequest) (int64, error)
	GetActivity(ctx context.Context, orgID int32, id int64) (*spec.OutreachActivityDetail, error)
	UpdateActivity(ctx context.Context, req *spec.UpdateActivityRequest) error
	CompleteActivity(ctx context.Context, orgID int32, id int64) error
	DeleteActivity(ctx context.Context, orgID int32, id int64) error

	// Engagement & Prospects Methods
	GetCampaignRecipients(ctx context.Context, orgID, campaignID int32) ([]*spec.CampaignRecipient, error)
	GetCampaignActivity(ctx context.Context, orgID, campaignID int32) ([]*spec.CampaignActivityEvent, error)
	GetProspects(ctx context.Context, orgID int32) ([]*spec.CampaignRecipient, error)
	GetProspectEngagement(ctx context.Context, orgID int32, leadID int64) (*spec.ProspectEngagementResponse, error)
	GetLeadOutreachActivity(ctx context.Context, orgID int32, leadID int64) ([]*spec.CampaignActivityEvent, error)

	GetProspectDetail(ctx context.Context, orgID int32, leadID int64) (*spec.ProspectDetailResponse, error)
	EnrollProspect(ctx context.Context, orgID int32, campaignID int32, leadID int64) error
	UpdateProspect(ctx context.Context, req *spec.UpdateProspectRequest) error
	PauseProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error
	ResumeProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error
	StopProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error
	GetFollowUps(ctx context.Context, orgID int32, filter string) ([]*spec.OutreachActivityDetail, error)
	CancelFollowUp(ctx context.Context, orgID int32, id int64) error
	RescheduleFollowUp(ctx context.Context, req *spec.RescheduleFollowUpRequest) error
}

type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) CreateCampaign(ctx context.Context, c *spec.Campaign) error {
	query := `
		INSERT INTO outreach_campaigns (org_id, name, status, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`
	res, err := d.db.ExecContext(ctx, query, c.OrgID, c.Name, c.Status)
	if err != nil {
		return fmt.Errorf("outreach.CreateCampaign: insert failed: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("outreach.CreateCampaign: get id failed: %w", err)
	}
	c.ID = int32(id)
	return nil
}

func (d *dataLayer) GetCampaignByID(ctx context.Context, orgID, id int32) (*spec.Campaign, error) {
	c := &spec.Campaign{}
	query := `
		SELECT id, org_id, name, status, created_at, updated_at
		FROM outreach_campaigns
		WHERE id = ? AND org_id = ?
	`
	err := d.db.GetContext(ctx, c, query, id, orgID)
	if err != nil {
		return nil, sql.ErrNoRows
	}

	// Fetch sequences
	seqQuery := `
		SELECT id, campaign_id, step_number, channel, template, delay_days, created_at, updated_at
		FROM outreach_sequences
		WHERE campaign_id = ?
		ORDER BY step_number ASC
	`
	err = d.db.SelectContext(ctx, &c.Sequences, seqQuery, c.ID)
	if err != nil {
		return nil, err
	}
	if c.Sequences == nil {
		c.Sequences = []spec.Sequence{}
	}

	return c, nil
}

func (d *dataLayer) ListCampaigns(ctx context.Context, orgID int32, limit, offset int) ([]*spec.Campaign, int, error) {
	var campaigns []*spec.Campaign
	var total int

	countQuery := `SELECT COUNT(*) FROM outreach_campaigns WHERE org_id = ?`
	if err := d.db.GetContext(ctx, &total, countQuery, orgID); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*spec.Campaign{}, 0, nil
	}

	listQuery := `
		SELECT id, org_id, name, status, created_at, updated_at
		FROM outreach_campaigns
		WHERE org_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	if err := d.db.SelectContext(ctx, &campaigns, listQuery, orgID, limit, offset); err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

func (d *dataLayer) UpdateCampaign(ctx context.Context, c *spec.Campaign) error {
	query := `
		UPDATE outreach_campaigns
		SET name = ?, status = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, c.Name, c.Status, c.ID, c.OrgID)
	return err
}

func (d *dataLayer) DeleteCampaign(ctx context.Context, orgID, id int32) error {
	query := `DELETE FROM outreach_campaigns WHERE id = ? AND org_id = ?`
	res, err := d.db.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *dataLayer) AddSequence(ctx context.Context, s *spec.Sequence) error {
	query := `
		INSERT INTO outreach_sequences (campaign_id, step_number, channel, template, delay_days, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	res, err := d.db.ExecContext(ctx, query, s.CampaignID, s.StepNumber, s.Channel, s.Template, s.DelayDays)
	if err != nil {
		return fmt.Errorf("outreach.AddSequence: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("outreach.AddSequence get id: %w", err)
	}
	s.ID = int32(id)
	return nil
}

func (d *dataLayer) GetCampaignSequence(ctx context.Context, orgID, campaignID int32) ([]*spec.Sequence, error) {
	var steps []*spec.Sequence
	query := `
		SELECT id, campaign_id, step_number, channel, COALESCE(name, '') as name,
		       COALESCE(subject, '') as subject, COALESCE(body, '') as body,
		       template, delay_days, created_at, updated_at
		FROM outreach_sequences
		WHERE campaign_id = ?
		ORDER BY step_number ASC
	`
	err := d.db.SelectContext(ctx, &steps, query, campaignID)
	if err != nil {
		return nil, err
	}
	if steps == nil {
		steps = []*spec.Sequence{}
	}
	return steps, nil
}

func (d *dataLayer) AddSequenceStep(ctx context.Context, s *spec.Sequence) error {
	// Find next step_number
	var maxStep sql.NullInt64
	err := d.db.GetContext(ctx, &maxStep, "SELECT MAX(step_number) FROM outreach_sequences WHERE campaign_id = ?", s.CampaignID)
	if err != nil {
		return err
	}
	s.StepNumber = 1
	if maxStep.Valid {
		s.StepNumber = int(maxStep.Int64) + 1
	}

	query := `
		INSERT INTO outreach_sequences (campaign_id, step_number, channel, name, subject, body, template, delay_days, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	res, err := d.db.ExecContext(ctx, query, s.CampaignID, s.StepNumber, s.Channel, s.Name, s.Subject, s.Body, s.Body, s.DelayDays)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	s.ID = int32(id)
	return nil
}

func (d *dataLayer) UpdateSequenceStep(ctx context.Context, s *spec.Sequence) error {
	query := `
		UPDATE outreach_sequences
		SET name = ?, subject = ?, body = ?, template = ?, delay_days = ?, updated_at = NOW()
		WHERE id = ? AND campaign_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, s.Name, s.Subject, s.Body, s.Body, s.DelayDays, s.ID, s.CampaignID)
	return err
}

func (d *dataLayer) DeleteSequenceStep(ctx context.Context, campaignID, stepID int32) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get step_number of step to delete
	var stepNum int
	err = tx.GetContext(ctx, &stepNum, "SELECT step_number FROM outreach_sequences WHERE id = ? AND campaign_id = ?", stepID, campaignID)
	if err != nil {
		return err
	}

	// Delete step
	_, err = tx.ExecContext(ctx, "DELETE FROM outreach_sequences WHERE id = ?", stepID)
	if err != nil {
		return err
	}

	// Reorder subsequent steps
	_, err = tx.ExecContext(ctx, "UPDATE outreach_sequences SET step_number = step_number - 1 WHERE campaign_id = ? AND step_number > ?", campaignID, stepNum)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *dataLayer) ReorderSequence(ctx context.Context, campaignID int32, stepIDs []int32) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for idx, stepID := range stepIDs {
		_, err = tx.ExecContext(ctx, "UPDATE outreach_sequences SET step_number = ? WHERE id = ? AND campaign_id = ?", idx+1, stepID, campaignID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *dataLayer) GetCampaignAudience(ctx context.Context, orgID, campaignID int32) ([]*spec.AudienceLead, error) {
	var leads []*spec.AudienceLead
	query := `
		SELECT l.id, l.company_name, l.contact_name, l.email, l.status, ocl.added_at
		FROM leads l
		JOIN outreach_campaign_leads ocl ON l.id = ocl.lead_id
		WHERE ocl.campaign_id = ? AND l.org_id = ?
		ORDER BY ocl.added_at DESC
	`
	err := d.db.SelectContext(ctx, &leads, query, campaignID, orgID)
	if err != nil {
		return nil, err
	}
	if leads == nil {
		leads = []*spec.AudienceLead{}
	}
	return leads, nil
}

func (d *dataLayer) AddCampaignAudience(ctx context.Context, campaignID int32, leadIDs []int32) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT IGNORE INTO outreach_campaign_leads (campaign_id, lead_id, added_at) VALUES (?, ?, NOW())`
	for _, leadID := range leadIDs {
		_, err := tx.ExecContext(ctx, query, campaignID, leadID)
		if err != nil {
			return err
		}
	}

	// Update campaign_id and source on leads
	updateQuery := `UPDATE leads SET campaign_id = ?, source = CASE WHEN source IS NULL OR source = '' THEN 'OUTREACH' ELSE source END WHERE id = ?`
	for _, leadID := range leadIDs {
		_, err := tx.ExecContext(ctx, updateQuery, campaignID, leadID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *dataLayer) RemoveCampaignAudience(ctx context.Context, campaignID, leadID int32) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM outreach_campaign_leads WHERE campaign_id = ? AND lead_id = ?", campaignID, leadID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE leads SET campaign_id = NULL WHERE id = ? AND campaign_id = ?", leadID, campaignID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *dataLayer) GetOutreachAnalytics(ctx context.Context, orgID int32) (*spec.OutreachDashboardResponse, error) {
	resp := &spec.OutreachDashboardResponse{
		UpcomingFollowups: []*spec.OutreachActivityDetail{},
		RecentActivities:  []*spec.OutreachActivityDetail{},
		OverdueItems:      []*spec.OutreachActivityDetail{},
	}

	// 1. Top Stats: Active Outreach
	_ = d.db.GetContext(ctx, &resp.ActiveOutreach,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND status IN ('PENDING', 'IN_PROGRESS')", orgID)

	// 2. Top Stats: Due Today
	_ = d.db.GetContext(ctx, &resp.DueToday,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND DATE(scheduled_at) = CURRENT_DATE() AND status IN ('PENDING', 'IN_PROGRESS')", orgID)

	// 3. Top Stats: Overdue
	_ = d.db.GetContext(ctx, &resp.Overdue,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND (status = 'OVERDUE' OR (scheduled_at < NOW() AND status IN ('PENDING', 'IN_PROGRESS')))", orgID)

	// 4. Top Stats: Completed
	_ = d.db.GetContext(ctx, &resp.Completed,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND status = 'COMPLETED'", orgID)

	// 5. Top Stats: Opportunities Created
	_ = d.db.GetContext(ctx, &resp.OpportunitiesCreated,
		"SELECT COUNT(*) FROM leads WHERE org_id = ? AND source = 'OUTREACH' AND status = 'CONVERTED'", orgID)

	// 5b. Operational stats
	_ = d.db.GetContext(ctx, &resp.ActiveCampaigns,
		"SELECT COUNT(*) FROM outreach_campaigns WHERE org_id = ? AND status = 'ACTIVE'", orgID)

	_ = d.db.GetContext(ctx, &resp.ActiveProspects,
		"SELECT COUNT(*) FROM outreach_campaign_leads ocl JOIN leads l ON ocl.lead_id = l.id WHERE l.org_id = ? AND ocl.status = 'ACTIVE'", orgID)

	_ = d.db.GetContext(ctx, &resp.EngagedProspects,
		"SELECT COUNT(*) FROM outreach_campaign_leads ocl JOIN leads l ON ocl.lead_id = l.id WHERE l.org_id = ? AND l.status = 'CONVERTED'", orgID)

	_ = d.db.GetContext(ctx, &resp.RepliesCount,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND activity_type = 'EMAIL' AND (subject LIKE '%Re:%' OR subject LIKE '%RE:%' OR description LIKE '%replied%')", orgID)

	// 6. Distribution by Type
	_ = d.db.GetContext(ctx, &resp.TypeEmail,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND activity_type = 'EMAIL'", orgID)
	_ = d.db.GetContext(ctx, &resp.TypeCall,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND activity_type = 'CALL'", orgID)
	_ = d.db.GetContext(ctx, &resp.TypeFollowup,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND activity_type = 'FOLLOW_UP'", orgID)
	_ = d.db.GetContext(ctx, &resp.TypeMeeting,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND activity_type = 'MEETING'", orgID)
	_ = d.db.GetContext(ctx, &resp.TypeOther,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND activity_type = 'OTHER'", orgID)

	// 7. Distribution by Status
	_ = d.db.GetContext(ctx, &resp.StatusPending,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND status = 'PENDING'", orgID)
	_ = d.db.GetContext(ctx, &resp.StatusInProgress,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND status = 'IN_PROGRESS'", orgID)
	_ = d.db.GetContext(ctx, &resp.StatusCompleted,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND status = 'COMPLETED'", orgID)
	_ = d.db.GetContext(ctx, &resp.StatusOverdue,
		"SELECT COUNT(*) FROM outreach_activities WHERE org_id = ? AND (status = 'OVERDUE' OR (scheduled_at < NOW() AND status IN ('PENDING', 'IN_PROGRESS')))", orgID)

	// 8. Upcoming Follow-ups (ordered by scheduled date, PENDING/IN_PROGRESS)
	err := d.db.SelectContext(ctx, &resp.UpcomingFollowups, `
		SELECT oa.*, l.company_name AS lead_company_name, l.contact_name AS lead_contact_name, l.email AS lead_email
		FROM outreach_activities oa
		LEFT JOIN leads l ON oa.lead_id = l.id
		WHERE oa.org_id = ? AND oa.status IN ('PENDING', 'IN_PROGRESS')
		ORDER BY oa.scheduled_at ASC
		LIMIT 10`, orgID)
	if err != nil {
		resp.UpcomingFollowups = []*spec.OutreachActivityDetail{}
	}

	// 9. Recent Activities
	err = d.db.SelectContext(ctx, &resp.RecentActivities, `
		SELECT oa.*, l.company_name AS lead_company_name, l.contact_name AS lead_contact_name, l.email AS lead_email, 
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))), ''), u.email) AS creator_name
		FROM outreach_activities oa
		LEFT JOIN leads l ON oa.lead_id = l.id
		LEFT JOIN users u ON oa.created_by = u.id
		WHERE oa.org_id = ?
		ORDER BY COALESCE(oa.scheduled_at, oa.created_at) DESC
		LIMIT 50`, orgID)
	if err != nil {
		resp.RecentActivities = []*spec.OutreachActivityDetail{}
	}

	// 10. Overdue Items (for Needs Attention panel)
	_ = d.db.SelectContext(ctx, &resp.OverdueItems, `
		SELECT oa.*, l.company_name AS lead_company_name, l.contact_name AS lead_contact_name, l.email AS lead_email, 
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))), ''), u.email) AS creator_name
		FROM outreach_activities oa
		LEFT JOIN leads l ON oa.lead_id = l.id
		LEFT JOIN users u ON oa.created_by = u.id
		WHERE oa.org_id = ? AND (oa.status = 'OVERDUE' OR (oa.scheduled_at < NOW() AND oa.status IN ('PENDING', 'IN_PROGRESS')))
		ORDER BY oa.scheduled_at ASC
		LIMIT 10`, orgID)
	if resp.OverdueItems == nil {
		resp.OverdueItems = []*spec.OutreachActivityDetail{}
	}

	return resp, nil
}

func (d *dataLayer) GetCampaignAnalytics(ctx context.Context, orgID, campaignID int32) (*spec.OutreachAnalyticsResponse, error) {
	resp := &spec.OutreachAnalyticsResponse{}

	// Verify campaign ownership
	var exists int
	err := d.db.GetContext(ctx, &exists, "SELECT COUNT(*) FROM outreach_campaigns WHERE id = ? AND org_id = ?", campaignID, orgID)
	if err != nil || exists == 0 {
		return nil, sql.ErrNoRows
	}

	err = d.db.GetContext(ctx, &resp.TotalRecipients, "SELECT COUNT(*) FROM outreach_campaign_leads WHERE campaign_id = ?", campaignID)
	if err != nil {
		return nil, err
	}

	_ = d.db.GetContext(ctx, &resp.CompletedRecipients, `
		SELECT COUNT(*) FROM leads l
		JOIN outreach_campaign_leads ocl ON l.id = ocl.lead_id
		WHERE ocl.campaign_id = ? AND l.status = 'CONVERTED'`, campaignID)

	_ = d.db.GetContext(ctx, &resp.ActiveRecipients, `
		SELECT COUNT(*) FROM leads l
		JOIN outreach_campaign_leads ocl ON l.id = ocl.lead_id
		WHERE ocl.campaign_id = ? AND l.status = 'QUALIFIED'`, campaignID)

	_ = d.db.GetContext(ctx, &resp.AwaitingRecipients, `
		SELECT COUNT(*) FROM leads l
		JOIN outreach_campaign_leads ocl ON l.id = ocl.lead_id
		WHERE ocl.campaign_id = ? AND l.status = 'NEW'`, campaignID)

	err = d.db.GetContext(ctx, &resp.LeadsGenerated, "SELECT COUNT(*) FROM leads WHERE org_id = ? AND campaign_id = ?", orgID, campaignID)
	if err != nil {
		return nil, err
	}

	if resp.TotalRecipients > 0 {
		resp.ConversionRate = float64(resp.LeadsGenerated) / float64(resp.TotalRecipients) * 100.0
	}

	var sent int
	_ = d.db.GetContext(ctx, &sent, `
		SELECT COUNT(*) FROM outreach_activities oa
		JOIN outreach_campaign_leads ocl ON oa.lead_id = ocl.lead_id
		WHERE ocl.campaign_id = ? AND oa.activity_type = 'EMAIL' AND oa.status = 'COMPLETED'`, campaignID)
	resp.EmailsSent = &sent

	return resp, nil
}

func (d *dataLayer) GetCampaignLeads(ctx context.Context, orgID, campaignID int32) ([]*spec.GeneratedLead, error) {
	var leads []*spec.GeneratedLead
	query := `
		SELECT l.id, l.company_name, l.contact_name, l.email, l.status,
		       COALESCE(l.converted_from_outreach_at, l.created_at) as converted_at,
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))), ''), u.email) as assigned_to_name
		FROM leads l
		LEFT JOIN users u ON l.assigned_to = u.id
		WHERE l.org_id = ? AND l.campaign_id = ?
		ORDER BY converted_at DESC
	`
	err := d.db.SelectContext(ctx, &leads, query, orgID, campaignID)
	if err != nil {
		return nil, err
	}
	if leads == nil {
		leads = []*spec.GeneratedLead{}
	}
	return leads, nil
}

func (d *dataLayer) GetCampaignInsights(ctx context.Context, orgID, campaignID int32) ([]spec.CampaignInsight, error) {
	var insights []spec.CampaignInsight

	// Check if sequence steps exist
	var seqCount int
	_ = d.db.GetContext(ctx, &seqCount, "SELECT COUNT(*) FROM outreach_sequences WHERE campaign_id = ?", campaignID)
	if seqCount == 0 {
		insights = append(insights, spec.CampaignInsight{
			Title:       "No Sequence Steps Configured",
			Description: "This campaign is active or draft, but has no email sequence steps set up. Contacts won't receive emails.",
			Severity:    "CRITICAL",
			Action:      "Add Sequence Step",
		})
	}

	// Check if audience has leads
	var audCount int
	_ = d.db.GetContext(ctx, &audCount, "SELECT COUNT(*) FROM outreach_campaign_leads WHERE campaign_id = ?", campaignID)
	if audCount == 0 {
		insights = append(insights, spec.CampaignInsight{
			Title:       "Audience List is Empty",
			Description: "You haven't added any leads to this outreach campaign. Add target contacts to begin email sequencing.",
			Severity:    "WARNING",
			Action:      "Add Leads",
		})
	}

	// Check if paused
	var status string
	_ = d.db.GetContext(ctx, &status, "SELECT status FROM outreach_campaigns WHERE id = ?", campaignID)
	if status == "PAUSED" {
		insights = append(insights, spec.CampaignInsight{
			Title:       "Campaign is Paused",
			Description: "Sequencing is currently suspended. Resume the campaign to restart email delivery to pending recipients.",
			Severity:    "INFO",
			Action:      "Launch Campaign",
		})
	}

	// Check if campaign generated leads but none converted to customers
	var leadsGen int
	_ = d.db.GetContext(ctx, &leadsGen, "SELECT COUNT(*) FROM leads WHERE campaign_id = ?", campaignID)
	if leadsGen > 0 {
		var customerCount int
		_ = d.db.GetContext(ctx, &customerCount, `
			SELECT COUNT(DISTINCT cll.customer_id) 
			FROM customer_lead_links cll
			JOIN leads l ON cll.lead_id = l.id
			WHERE l.campaign_id = ?`, campaignID)
		if customerCount == 0 {
			insights = append(insights, spec.CampaignInsight{
				Title:       "Unconverted Lead Pipeline",
				Description: "This campaign generated qualified leads, but none have converted to active customer accounts. Follow up to secure orders.",
				Severity:    "INFO",
				Action:      "Review Leads",
			})
		}
	}

	if insights == nil {
		insights = []spec.CampaignInsight{}
	}

	return insights, nil
}

func (d *dataLayer) GetConversionFunnel(ctx context.Context, orgID int32) (*spec.ConversionFunnelResponse, error) {
	var stages []spec.FunnelStage

	// Stage 1: Campaign Audience
	var count1 int
	err := d.db.GetContext(ctx, &count1, `
		SELECT COUNT(DISTINCT ocl.lead_id) 
		FROM outreach_campaign_leads ocl
		JOIN outreach_campaigns oc ON ocl.campaign_id = oc.id
		WHERE oc.org_id = ?`, orgID)
	if err != nil {
		return nil, err
	}
	stages = append(stages, spec.FunnelStage{Stage: "Campaign Audience", Count: count1})

	// Stage 2: Recipients Contacted
	var count2 int
	err = d.db.GetContext(ctx, &count2, `
		SELECT COUNT(DISTINCT ocl.lead_id) 
		FROM outreach_campaign_leads ocl
		JOIN outreach_campaigns oc ON ocl.campaign_id = oc.id
		JOIN leads l ON ocl.lead_id = l.id
		WHERE oc.org_id = ? AND l.status IN ('QUALIFIED', 'CONVERTED', 'REJECTED')`, orgID)
	if err != nil {
		return nil, err
	}
	stages = append(stages, spec.FunnelStage{Stage: "Recipients Contacted", Count: count2})

	// Stage 3: Qualified Prospects
	var count3 int
	err = d.db.GetContext(ctx, &count3, `
		SELECT COUNT(DISTINCT ocl.lead_id) 
		FROM outreach_campaign_leads ocl
		JOIN outreach_campaigns oc ON ocl.campaign_id = oc.id
		JOIN leads l ON ocl.lead_id = l.id
		WHERE oc.org_id = ? AND l.status IN ('QUALIFIED', 'CONVERTED')`, orgID)
	if err != nil {
		return nil, err
	}
	stages = append(stages, spec.FunnelStage{Stage: "Qualified Prospects", Count: count3})

	// Stage 4: Leads Generated
	var count4 int
	err = d.db.GetContext(ctx, &count4, "SELECT COUNT(*) FROM leads WHERE org_id = ? AND source = 'OUTREACH'", orgID)
	if err != nil {
		return nil, err
	}
	stages = append(stages, spec.FunnelStage{Stage: "Leads Generated", Count: count4})

	// Stage 5: Customers
	var count5 int
	err = d.db.GetContext(ctx, &count5, `
		SELECT COUNT(DISTINCT cll.customer_id) 
		FROM customer_lead_links cll
		JOIN leads l ON cll.lead_id = l.id
		WHERE l.org_id = ? AND l.source = 'OUTREACH'`, orgID)
	if err != nil {
		return nil, err
	}
	stages = append(stages, spec.FunnelStage{Stage: "Converted Customers", Count: count5})

	return &spec.ConversionFunnelResponse{Stages: stages}, nil
}

func (d *dataLayer) CreateActivity(ctx context.Context, req *spec.CreateActivityRequest) (int64, error) {
	query := `INSERT INTO outreach_activities (org_id, lead_id, customer_id, activity_type, subject, description, status, priority, scheduled_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := d.db.ExecContext(ctx, query,
		req.OrgID,
		req.LeadID,
		req.CustomerID,
		req.ActivityType,
		req.Subject,
		req.Description,
		req.Status,
		req.Priority,
		req.ScheduledAt,
		req.CreatedBy,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *dataLayer) GetActivity(ctx context.Context, orgID int32, id int64) (*spec.OutreachActivityDetail, error) {
	var act spec.OutreachActivityDetail
	query := `SELECT oa.*, l.company_name AS lead_company_name, l.contact_name AS lead_contact_name, l.email AS lead_email, 
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))), ''), u.email) AS creator_name
		FROM outreach_activities oa
		LEFT JOIN leads l ON oa.lead_id = l.id
		LEFT JOIN users u ON oa.created_by = u.id
		WHERE oa.org_id = ? AND oa.id = ?`
	err := d.db.GetContext(ctx, &act, query, orgID, id)
	if err != nil {
		return nil, err
	}
	return &act, nil
}

func (d *dataLayer) UpdateActivity(ctx context.Context, req *spec.UpdateActivityRequest) error {
	query := `UPDATE outreach_activities 
		SET lead_id = ?, customer_id = ?, activity_type = ?, subject = ?, description = ?, status = ?, priority = ?, scheduled_at = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query,
		req.LeadID,
		req.CustomerID,
		req.ActivityType,
		req.Subject,
		req.Description,
		req.Status,
		req.Priority,
		req.ScheduledAt,
		req.OrgID,
		req.ID,
	)
	return err
}

func (d *dataLayer) CompleteActivity(ctx context.Context, orgID int32, id int64) error {
	query := `UPDATE outreach_activities 
		SET status = 'COMPLETED', completed_at = NOW(), updated_at = NOW()
		WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, orgID, id)
	return err
}

func (d *dataLayer) DeleteActivity(ctx context.Context, orgID int32, id int64) error {
	query := `DELETE FROM outreach_activities WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, orgID, id)
	return err
}

func (d *dataLayer) GetCampaignRecipients(ctx context.Context, orgID, campaignID int32) ([]*spec.CampaignRecipient, error) {
	var recipients []*spec.CampaignRecipient
	query := `
		SELECT l.id AS lead_id, l.company_name, l.contact_name, l.email, l.status AS lead_status
		FROM leads l
		JOIN outreach_campaign_leads ocl ON l.id = ocl.lead_id
		WHERE ocl.campaign_id = ? AND l.org_id = ?
		ORDER BY ocl.added_at DESC
	`
	err := d.db.SelectContext(ctx, &recipients, query, campaignID, orgID)
	if err != nil {
		return nil, err
	}
	if recipients == nil {
		recipients = []*spec.CampaignRecipient{}
	}

	for _, r := range recipients {
		var sent int
		_ = d.db.GetContext(ctx, &sent, "SELECT COUNT(*) FROM outreach_activities WHERE lead_id = ? AND status = 'COMPLETED' AND activity_type = 'EMAIL'", r.LeadID)
		r.EmailsSent = sent

		var completedCount int
		_ = d.db.GetContext(ctx, &completedCount, "SELECT COUNT(*) FROM outreach_activities WHERE lead_id = ? AND status = 'COMPLETED'", r.LeadID)
		
		if r.LeadStatus == "CONVERTED" {
			r.EngagementStatus = "ENGAGED"
		} else if completedCount > 0 {
			r.EngagementStatus = "CONTACTED"
		} else {
			r.EngagementStatus = "NOT_CONTACTED"
		}

		type lastActStruct struct {
			Subject     string     `db:"subject"`
			CompletedAt *time.Time `db:"completed_at"`
			CreatedAt   time.Time  `db:"created_at"`
		}
		var lastAct lastActStruct
		err = d.db.GetContext(ctx, &lastAct, "SELECT subject, completed_at, created_at FROM outreach_activities WHERE lead_id = ? ORDER BY created_at DESC LIMIT 1", r.LeadID)
		if err == nil {
			if lastAct.CompletedAt != nil {
				r.LastActivityAt = lastAct.CompletedAt
			} else {
				r.LastActivityAt = &lastAct.CreatedAt
			}
			r.LastActivityDesc = &lastAct.Subject
		}
	}

	return recipients, nil
}

func (d *dataLayer) GetCampaignActivity(ctx context.Context, orgID, campaignID int32) ([]*spec.CampaignActivityEvent, error) {
	var events []*spec.CampaignActivityEvent
	query := `
		SELECT oa.id, oa.activity_type, oa.subject, oa.description, oa.status, oa.scheduled_at, oa.completed_at, oa.created_at,
		       l.company_name AS lead_company_name, l.contact_name AS lead_contact_name,
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))), ''), u.email) AS creator_name
		FROM outreach_activities oa
		JOIN outreach_campaign_leads ocl ON oa.lead_id = ocl.lead_id
		LEFT JOIN leads l ON oa.lead_id = l.id
		LEFT JOIN users u ON oa.created_by = u.id
		WHERE ocl.campaign_id = ? AND oa.org_id = ?
		ORDER BY oa.created_at DESC
	`
	err := d.db.SelectContext(ctx, &events, query, campaignID, orgID)
	if err != nil {
		return nil, err
	}
	if events == nil {
		events = []*spec.CampaignActivityEvent{}
	}
	return events, nil
}

func (d *dataLayer) GetProspects(ctx context.Context, orgID int32) ([]*spec.CampaignRecipient, error) {
	var recipients []*spec.CampaignRecipient
	query := `
		SELECT 
			l.id AS lead_id, 
			l.company_name, 
			l.contact_name, 
			l.email, 
			l.phone,
			l.status AS lead_status,
			ocl.campaign_id,
			oc.name AS campaign_name,
			ocl.status,
			ocl.current_step,
			ocl.next_scheduled_at,
			ocl.last_activity_at
		FROM leads l
		JOIN outreach_campaign_leads ocl ON l.id = ocl.lead_id
		LEFT JOIN outreach_campaigns oc ON ocl.campaign_id = oc.id
		WHERE l.org_id = ?
		ORDER BY ocl.added_at DESC
	`
	err := d.db.SelectContext(ctx, &recipients, query, orgID)
	if err != nil {
		return nil, err
	}
	if recipients == nil {
		recipients = []*spec.CampaignRecipient{}
	}

	for _, r := range recipients {
		var sent int
		_ = d.db.GetContext(ctx, &sent, "SELECT COUNT(*) FROM outreach_activities WHERE lead_id = ? AND status = 'COMPLETED' AND activity_type = 'EMAIL'", r.LeadID)
		r.EmailsSent = sent

		var completedCount int
		_ = d.db.GetContext(ctx, &completedCount, "SELECT COUNT(*) FROM outreach_activities WHERE lead_id = ? AND status = 'COMPLETED'", r.LeadID)
		
		if r.LeadStatus == "CONVERTED" {
			r.EngagementStatus = "ENGAGED"
		} else if completedCount > 0 {
			r.EngagementStatus = "CONTACTED"
		} else {
			r.EngagementStatus = "NOT_CONTACTED"
		}

		type lastActStruct struct {
			Subject     string     `db:"subject"`
			CompletedAt *time.Time `db:"completed_at"`
			CreatedAt   time.Time  `db:"created_at"`
		}
		var lastAct lastActStruct
		err = d.db.GetContext(ctx, &lastAct, "SELECT subject, completed_at, created_at FROM outreach_activities WHERE lead_id = ? ORDER BY created_at DESC LIMIT 1", r.LeadID)
		if err == nil {
			if lastAct.CompletedAt != nil {
				r.LastActivityAt = lastAct.CompletedAt
			} else {
				r.LastActivityAt = &lastAct.CreatedAt
			}
			r.LastActivityDesc = &lastAct.Subject
		}
	}
	return recipients, nil
}

func (d *dataLayer) GetProspectEngagement(ctx context.Context, orgID int32, leadID int64) (*spec.ProspectEngagementResponse, error) {
	var lead struct {
		ID          int64   `db:"id"`
		CompanyName string  `db:"company_name"`
		ContactName *string `db:"contact_name"`
		Email       *string `db:"email"`
		Status      string  `db:"status"`
	}
	err := d.db.GetContext(ctx, &lead, "SELECT id, company_name, contact_name, email, status FROM leads WHERE id = ? AND org_id = ?", leadID, orgID)
	if err != nil {
		return nil, err
	}

	var campaignName *string
	_ = d.db.GetContext(ctx, &campaignName, `
		SELECT oc.name FROM outreach_campaigns oc
		JOIN outreach_campaign_leads ocl ON oc.id = ocl.campaign_id
		WHERE ocl.lead_id = ? LIMIT 1`, leadID)

	var sent int
	_ = d.db.GetContext(ctx, &sent, "SELECT COUNT(*) FROM outreach_activities WHERE lead_id = ? AND status = 'COMPLETED' AND activity_type = 'EMAIL'", leadID)

	var completedCount int
	_ = d.db.GetContext(ctx, &completedCount, "SELECT COUNT(*) FROM outreach_activities WHERE lead_id = ? AND status = 'COMPLETED'", leadID)
	
	engagementStatus := "NOT_CONTACTED"
	if lead.Status == "CONVERTED" {
		engagementStatus = "ENGAGED"
	} else if completedCount > 0 {
		engagementStatus = "CONTACTED"
	}

	var timeline []*spec.CampaignActivityEvent
	timelineQuery := `
		SELECT oa.id, oa.activity_type, oa.subject, oa.description, oa.status, oa.scheduled_at, oa.completed_at, oa.created_at,
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))), ''), u.email) AS creator_name
		FROM outreach_activities oa
		LEFT JOIN users u ON oa.created_by = u.id
		WHERE oa.lead_id = ? AND oa.org_id = ?
		ORDER BY oa.created_at DESC
	`
	_ = d.db.SelectContext(ctx, &timeline, timelineQuery, leadID, orgID)
	if timeline == nil {
		timeline = []*spec.CampaignActivityEvent{}
	}

	var lastActivityAt *time.Time
	if len(timeline) > 0 {
		if timeline[0].CompletedAt != nil {
			lastActivityAt = timeline[0].CompletedAt
		} else {
			lastActivityAt = &timeline[0].CreatedAt
		}
	}

	return &spec.ProspectEngagementResponse{
		LeadID:           lead.ID,
		CompanyName:      lead.CompanyName,
		ContactName:      lead.ContactName,
		Email:            lead.Email,
		LeadStatus:       lead.Status,
		EngagementStatus: engagementStatus,
		CampaignName:     campaignName,
		EmailsSent:       sent,
		LastActivityAt:   lastActivityAt,
		Timeline:         timeline,
	}, nil
}

func (d *dataLayer) GetLeadOutreachActivity(ctx context.Context, orgID int32, leadID int64) ([]*spec.CampaignActivityEvent, error) {
	var timeline []*spec.CampaignActivityEvent
	query := `
		SELECT oa.id, oa.activity_type, oa.subject, oa.description, oa.status, oa.scheduled_at, oa.completed_at, oa.created_at,
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))), ''), u.email) AS creator_name
		FROM outreach_activities oa
		LEFT JOIN users u ON oa.created_by = u.id
		WHERE oa.lead_id = ? AND oa.org_id = ?
		ORDER BY oa.created_at DESC
	`
	err := d.db.SelectContext(ctx, &timeline, query, leadID, orgID)
	if err != nil {
		return nil, err
	}
	if timeline == nil {
		timeline = []*spec.CampaignActivityEvent{}
	}
	return timeline, nil
}

func (d *dataLayer) GetProspectDetail(ctx context.Context, orgID int32, leadID int64) (*spec.ProspectDetailResponse, error) {
	var prospect spec.CampaignRecipient
	query := `
		SELECT ocl.lead_id, l.company_name, l.contact_name, l.email, l.phone, l.status AS lead_status,
		       ocl.campaign_id, oc.name AS campaign_name, ocl.status, ocl.current_step,
		       ocl.last_activity_at, ocl.next_scheduled_at, 
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))), ''), u.email) AS owner_name
		FROM outreach_campaign_leads ocl
		JOIN leads l ON ocl.lead_id = l.id
		LEFT JOIN outreach_campaigns oc ON ocl.campaign_id = oc.id
		LEFT JOIN users u ON l.assigned_to = u.id
		WHERE l.org_id = ? AND ocl.lead_id = ?
		LIMIT 1
	`
	err := d.db.GetContext(ctx, &prospect, query, orgID, leadID)
	if err != nil {
		return nil, err
	}

	var sent int
	_ = d.db.GetContext(ctx, &sent, "SELECT COUNT(*) FROM outreach_activities WHERE lead_id = ? AND status = 'COMPLETED' AND activity_type = 'EMAIL'", leadID)
	prospect.EmailsSent = sent

	var completedCount int
	_ = d.db.GetContext(ctx, &completedCount, "SELECT COUNT(*) FROM outreach_activities WHERE lead_id = ? AND status = 'COMPLETED'", leadID)
	if prospect.LeadStatus == "CONVERTED" {
		prospect.EngagementStatus = "ENGAGED"
	} else if completedCount > 0 {
		prospect.EngagementStatus = "CONTACTED"
	} else {
		prospect.EngagementStatus = "NOT_CONTACTED"
	}

	type lastActStruct struct {
		Subject     string     `db:"subject"`
		CompletedAt *time.Time `db:"completed_at"`
		CreatedAt   time.Time  `db:"created_at"`
	}
	var lastAct lastActStruct
	err = d.db.GetContext(ctx, &lastAct, "SELECT subject, completed_at, created_at FROM outreach_activities WHERE lead_id = ? ORDER BY created_at DESC LIMIT 1", leadID)
	if err == nil {
		if lastAct.CompletedAt != nil {
			prospect.LastActivityAt = lastAct.CompletedAt
		} else {
			prospect.LastActivityAt = &lastAct.CreatedAt
		}
		prospect.LastActivityDesc = &lastAct.Subject
	}

	// Fetch involving campaigns
	var campaigns []*spec.Campaign
	campaignsQuery := `
		SELECT oc.* FROM outreach_campaigns oc
		JOIN outreach_campaign_leads ocl ON oc.id = ocl.campaign_id
		WHERE ocl.lead_id = ? AND oc.org_id = ?
	`
	_ = d.db.SelectContext(ctx, &campaigns, campaignsQuery, leadID, orgID)
	if campaigns == nil {
		campaigns = []*spec.Campaign{}
	}

	// Fetch sequence steps
	var sequenceSteps []*spec.Sequence
	if prospect.CampaignID != nil {
		sequenceSteps, _ = d.GetCampaignSequence(ctx, orgID, int32(*prospect.CampaignID))
	}
	if sequenceSteps == nil {
		sequenceSteps = []*spec.Sequence{}
	}

	// Timeline activities
	activities, _ := d.GetLeadOutreachActivity(ctx, orgID, leadID)
	if activities == nil {
		activities = []*spec.CampaignActivityEvent{}
	}

	// Follow-ups (outreach activities list)
	var followUps []*spec.OutreachActivityDetail
	followUpsQuery := `
		SELECT oa.id, oa.org_id, oa.lead_id, oa.customer_id, oa.activity_type, oa.subject, oa.description, 
		       oa.status, oa.priority, oa.scheduled_at, oa.completed_at, oa.created_at, oa.created_by,
		       l.company_name AS lead_company_name, l.contact_name AS lead_contact_name,
		       COALESCE(NULLIF(TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))), ''), u.email) AS creator_name
		FROM outreach_activities oa
		LEFT JOIN leads l ON oa.lead_id = l.id
		LEFT JOIN users u ON oa.created_by = u.id
		WHERE oa.org_id = ? AND oa.lead_id = ?
		ORDER BY oa.scheduled_at ASC
	`
	_ = d.db.SelectContext(ctx, &followUps, followUpsQuery, orgID, leadID)
	if followUps == nil {
		followUps = []*spec.OutreachActivityDetail{}
	}

	return &spec.ProspectDetailResponse{
		Prospect:      &prospect,
		Campaigns:     campaigns,
		SequenceSteps: sequenceSteps,
		Activities:    activities,
		FollowUps:     followUps,
	}, nil
}

func (d *dataLayer) EnrollProspect(ctx context.Context, orgID int32, campaignID int32, leadID int64) error {
	// Verify lead ownership
	var exists int
	err := d.db.GetContext(ctx, &exists, "SELECT COUNT(*) FROM leads WHERE id = ? AND org_id = ?", leadID, orgID)
	if err != nil || exists == 0 {
		return sql.ErrNoRows
	}

	// Enroll into campaign
	query := `INSERT IGNORE INTO outreach_campaign_leads (campaign_id, lead_id, status, current_step) VALUES (?, ?, 'ACTIVE', 1)`
	_, err = d.db.ExecContext(ctx, query, campaignID, leadID)
	return err
}

func (d *dataLayer) UpdateProspect(ctx context.Context, req *spec.UpdateProspectRequest) error {
	query := `
		UPDATE outreach_campaign_leads ocl
		JOIN leads l ON ocl.lead_id = l.id
		SET ocl.status = ?, ocl.current_step = ?
		WHERE l.org_id = ? AND ocl.lead_id = ? AND ocl.campaign_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, req.Status, req.CurrentStep, req.OrgID, req.LeadID, req.CampaignID)
	return err
}

func (d *dataLayer) PauseProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error {
	query := `
		UPDATE outreach_campaign_leads ocl
		JOIN leads l ON ocl.lead_id = l.id
		SET ocl.status = 'PAUSED'
		WHERE l.org_id = ? AND ocl.lead_id = ? AND ocl.campaign_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, orgID, leadID, campaignID)
	return err
}

func (d *dataLayer) ResumeProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error {
	query := `
		UPDATE outreach_campaign_leads ocl
		JOIN leads l ON ocl.lead_id = l.id
		SET ocl.status = 'ACTIVE'
		WHERE l.org_id = ? AND ocl.lead_id = ? AND ocl.campaign_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, orgID, leadID, campaignID)
	return err
}

func (d *dataLayer) StopProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error {
	query := `
		UPDATE outreach_campaign_leads ocl
		JOIN leads l ON ocl.lead_id = l.id
		SET ocl.status = 'COMPLETED'
		WHERE l.org_id = ? AND ocl.lead_id = ? AND ocl.campaign_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, orgID, leadID, campaignID)
	return err
}

func (d *dataLayer) GetFollowUps(ctx context.Context, orgID int32, filter string) ([]*spec.OutreachActivityDetail, error) {
	var activities []*spec.OutreachActivityDetail
	var query string
	var args []interface{}
	
	baseQuery := `
		SELECT oa.id, oa.org_id, oa.lead_id, oa.customer_id, oa.activity_type, oa.subject, oa.description, 
		       oa.status, oa.priority, oa.scheduled_at, oa.completed_at, oa.created_at, oa.created_by,
		       l.company_name AS lead_company_name, l.contact_name AS lead_contact_name,
		       c.name AS customer_company_name, c.contact_name AS customer_contact_name,
		       CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, '')) AS creator_name
		FROM outreach_activities oa
		LEFT JOIN leads l ON oa.lead_id = l.id
		LEFT JOIN customers c ON oa.customer_id = c.id
		LEFT JOIN users u ON oa.created_by = u.id
		WHERE oa.org_id = ?
	`
	args = append(args, orgID)
	
	switch filter {
	case "today":
		query = baseQuery + " AND oa.status = 'PENDING' AND DATE(oa.scheduled_at) = CURDATE() ORDER BY oa.scheduled_at ASC"
	case "upcoming":
		query = baseQuery + " AND oa.status = 'PENDING' AND oa.scheduled_at > NOW() ORDER BY oa.scheduled_at ASC"
	case "overdue":
		query = baseQuery + " AND oa.status = 'PENDING' AND oa.scheduled_at < NOW() ORDER BY oa.scheduled_at ASC"
	case "completed":
		query = baseQuery + " AND oa.status = 'COMPLETED' ORDER BY oa.completed_at DESC"
	default:
		query = baseQuery + " ORDER BY oa.scheduled_at ASC"
	}
	
	err := d.db.SelectContext(ctx, &activities, query, args...)
	if err != nil {
		return nil, err
	}
	if activities == nil {
		activities = []*spec.OutreachActivityDetail{}
	}
	return activities, nil
}

func (d *dataLayer) CancelFollowUp(ctx context.Context, orgID int32, id int64) error {
	query := `UPDATE outreach_activities SET status = 'CANCELLED' WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, orgID, id)
	return err
}

func (d *dataLayer) RescheduleFollowUp(ctx context.Context, req *spec.RescheduleFollowUpRequest) error {
	query := `UPDATE outreach_activities SET scheduled_at = ?, status = 'PENDING' WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, req.ScheduledAt, req.OrgID, req.ID)
	return err
}

