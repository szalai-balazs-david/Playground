from flask import Flask
from flask_sqlalchemy import SQLAlchemy

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://postgres:6%R4tC%3Ixbh0#Au@localhost:5432/sample'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

db = SQLAlchemy(app)

class User(db.Model):
  __tablename__ = 'users'
  id = db.Column(db.Integer, primary_key=True, autoincrement=True)
  name = db.Column(db.String(), nullable=False)

  def __repr__(self):
      return f'<User {self.id}, {self.name}>'
db.drop_all()
db.create_all()

user1 = User(name='Bob')
user3 = User(name='Rebecca')
user2 = User(name='Nicole')
user4 = User(name='Boris')
user5 = User(name='Bob')

db.session.add(user1)
db.session.add(user2)
db.session.add(user3)
db.session.add(user4)
db.session.add(user5)
db.session.commit()

@app.route('/')
def index():
    displayString = ''
    displayString = displayString + 'All the Bobs:\n'
    bobs = db.session.query(User).filter(User.name=='Bob').all()
    for bob in bobs:
        displayString = displayString + str(bob.id) + ': ' + bob.name + '\n'

    displayString = displayString + 'Contains "b":\n'
    hasB = db.session.query(User).filter(User.name.contains('b')).all()
    for b in hasB:
        displayString = displayString + str(b.id) + ': ' + b.name + '\n'

    displayString = displayString + 'First 2 that contains "b":\n'
    hasB2 = db.session.query(User).filter(User.name.contains('b')).limit(2)
    for b in hasB2:
        displayString = displayString + str(b.id) + ': ' + b.name + '\n'

    displayString = displayString + 'Contains "b" - case insensitive:\n'
    hasInsensitiveB = db.session.query(User).filter(User.name.ilike('%b%')).all()
    for b in hasInsensitiveB:
        displayString = displayString + str(b.id) + ': ' + b.name + '\n'

    displayString = displayString + 'Number of Bobs: ' + str(db.session.query(User).filter(User.name=='Bob').count())

    return displayString

if __name__ == '__main__':
    app.debug = True
    app.run()