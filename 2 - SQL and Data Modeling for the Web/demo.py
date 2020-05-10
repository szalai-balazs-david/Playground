import psycopg2

connection = psycopg2.connect(database="sample", user="postgres", password="6%R4tC%3Ixbh0#Au")

cursor = connection.cursor()

cursor.execute('DROP TABLE IF EXISTS table2;')

cursor.execute('''CREATE TABLE table2 (id INTEGER PRIMARY KEY, completed BOOLEAN NOT NULL DEFAULT False);''')

cursor.execute('INSERT INTO table2 (id, completed) VALUES (%s, %s);', (1, True))

SQL = 'INSERT INTO table2 (id, completed) VALUES (%(id)s, %(completed)s);'

data = {
  'id': 2,
  'completed': True
}
cursor.execute(SQL, data)

connection.commit()

cursor.execute('SELECT * from table2;')
result = cursor.fetchone()
print(result)
result = cursor.fetchone()
print(result)
result = cursor.fetchone()
print(result)

cursor.close()
connection.close()