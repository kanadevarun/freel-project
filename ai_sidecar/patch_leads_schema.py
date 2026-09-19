import pymysql

conn = pymysql.connect(host='127.0.0.1', port=3306, user='root', password='', db='freel_mysql', autocommit=True)
with conn.cursor() as cur:
    # 1. Fix ai_score on leads
    cur.execute("UPDATE leads SET ai_score = 0 WHERE ai_score IS NULL")
    cur.execute("ALTER TABLE leads MODIFY COLUMN ai_score INT NOT NULL DEFAULT 0")
    print("Fixed leads.ai_score.")
    
    # 2. Fix columns on lead_interactions
    cur.execute("DESCRIBE lead_interactions")
    existing_cols = [r[0] for r in cur.fetchall()]
    print("Existing lead_interactions cols:", existing_cols)
    
    col_defs = [
        ("channel", "VARCHAR(50) DEFAULT 'EMAIL'"),
        ("subject", "VARCHAR(255) DEFAULT ''"),
        ("content", "LONGTEXT"),
        ("raw_email_id", "VARCHAR(255) DEFAULT ''"),
        ("thread_id", "VARCHAR(255) DEFAULT ''"),
        ("sentiment", "VARCHAR(50) DEFAULT 'NEUTRAL'"),
        ("intent", "VARCHAR(50) DEFAULT 'QUESTION'"),
        ("linked_rfq_id", "BIGINT DEFAULT NULL"),
        ("ai_confidence", "INT DEFAULT 0"),
        ("ai_summary", "TEXT"),
        ("drafted_reply", "LONGTEXT"),
        ("parent_interaction_id", "BIGINT DEFAULT NULL")
    ]
    for col_name, col_type in col_defs:
        if col_name not in existing_cols:
            cur.execute(f"ALTER TABLE lead_interactions ADD COLUMN {col_name} {col_type}")
            print(f"Added column {col_name}")
            
    # Copy full_body into content if content is null
    cur.execute("UPDATE lead_interactions SET content = full_body WHERE content IS NULL")
    cur.execute("UPDATE lead_interactions SET subject = summary WHERE subject = '' OR subject IS NULL")
    print("lead_interactions aligned successfully!")
conn.close()
