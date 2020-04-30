from flask import Flask, render_template, request, redirect, url_for, jsonify, abort
from flask_sqlalchemy import SQLAlchemy
from flask_migrate import Migrate
import sys

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://postgres:6%R4tC%3Ixbh0#Au@localhost:5432/todoapp'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
db = SQLAlchemy(app)
migrate = Migrate(app, db)

class Todo(db.Model):
  __tablename__ = 'todos'
  id = db.Column(db.Integer, primary_key=True, autoincrement=True)
  description = db.Column(db.String(), nullable=False)
  completed = db.Column(db.Boolean, nullable=False, default = False)

  def __repr__(self):
      return f'<Todo {self.id} {self.name}>'

@app.route('/')
def index():
    return render_template('index.html', data=Todo.query.order_by('id').all())

@app.route('/todos/create', methods=['POST'])
def create():
    error = False
    body = {}
    try:
        desc = request.get_json()['description']
        todo = Todo(description=desc)
        db.session.add(todo)
        db.session.commit()
        body['description'] = todo.description
    except:
        error=True
        db.session.rollback()
        print(sys.exc_into())
    finally:
        db.session.close()
    if error:
        abort(400)
    else:
        return jsonify(body)

@app.route('/todos/<todo_id>/set-completed', methods=['POST'])
def set_completed_todo(todo_id):
    try:
        completed = request.get_json()['completed']
        todo = Todo.query.get(todo_id)
        todo.completed = completed
        db.session.commit()
    except:
        db.session.rollback()
    finally:
        db.session.close()
    return redirect(url_for('index'))

@app.route('/todos/<todo_id>/delete', methods=['GET'])
def delete_todo(todo_id):
    try:
        todo = Todo.query.get(todo_id)
        db.session.delete(todo)
        db.session.commit()
    except:
        db.session.rollback()
    finally:
        db.session.close()
    return redirect(url_for('index'))


if __name__ == '__main__':
    app.debug = True
    app.run()