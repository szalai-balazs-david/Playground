from flask import Flask, render_template, request, redirect, url_for, jsonify
from flask_sqlalchemy import SQLAlchemy

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://postgres:6%R4tC%3Ixbh0#Au@localhost:5432/todoapp'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
db = SQLAlchemy(app)

class Todo(db.Model):
  __tablename__ = 'todos'
  id = db.Column(db.Integer, primary_key=True, autoincrement=True)
  description = db.Column(db.String(), nullable=False)

  def __repr__(self):
      return f'<Todo {self.id} {self.name}>'

db.drop_all()
db.create_all()

user1 = Todo(description='ToDo1')
user3 = Todo(description='ToDo2')
user2 = Todo(description='ToDo3')
user4 = Todo(description='ToDo4')
user5 = Todo(description='ToDo5')

db.session.add(user1)
db.session.add(user2)
db.session.add(user3)
db.session.add(user4)
db.session.add(user5)
db.session.commit()

@app.route('/')
def index():
    return render_template('index.html', data=Todo.query.all())

@app.route('/todos/create', methods=['POST'])
def create():
    desc = request.get_json()['description']
    todo = Todo(description=desc)
    db.session.add(todo)
    db.session.commit()
    return jsonify({
        'description': todo.description
    })

if __name__ == '__main__':
    app.debug = True
    app.run()