package notifications

type Notification struct {
	ID        int32  `json:"id" db:"id"`
	OrgID     int32  `json:"org_id" db:"org_id"`
	Title     string `json:"title" db:"title"`
	Message   string `json:"message" db:"message"`
	Type      string `json:"type" db:"type"` // e.g. "INFO", "SUCCESS", "WARNING", "ERROR"
	IsRead    bool   `json:"is_read" db:"is_read"`
	CreatedAt string `json:"created_at" db:"created_at"`
}
