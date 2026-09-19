import pymysql
import sys

def check_db():
    conn = pymysql.connect(
        host='127.0.0.1',
        port=3306,
        user='root',
        password='',
        database='freel_mysql',
        cursorclass=pymysql.cursors.DictCursor
    )
    with conn.cursor() as cursor:
        cursor.execute("SHOW TABLES")
        all_tables = [list(row.values())[0] for row in cursor.fetchall()]
        print(f"Total tables in freel_mysql: {len(all_tables)}")
        
        keywords = ["predict", "track", "feed", "gov", "violat", "audit", "ship", "inv", "cust", "lead", "rfq", "quote", "approv"]
        for kw in keywords:
            matching = [t for t in all_tables if kw in t]
            print(f"\nKeyword '{kw}': {matching}")
            for t in matching:
                cursor.execute(f"SELECT COUNT(*) as cnt FROM {t}")
                print(f"  {t}: {cursor.fetchone()['cnt']} rows")
    conn.close()

if __name__ == '__main__':
    check_db()
