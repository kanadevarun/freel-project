import pymysql

conn = pymysql.connect(host='127.0.0.1', user='root', password='', database='freel_mysql')
cur = conn.cursor(pymysql.cursors.DictCursor)

cur.execute("SHOW TABLES")
tables = [list(r.values())[0] for r in cur.fetchall()]
print("Matching tables:")
for t in tables:
    if any(k in t.lower() for k in ['sub', 'plan', 'tier', 'usage', 'limit']):
        print(" ", t)

conn.close()
