from flask import Flask, request, Response, make_response, session
import requests
import json
import argparse
import logging
import sys

from flask_cors import CORS

# email
from flask_mail import Mail

# database import
from tinydb import TinyDB, where, Query
db = TinyDB('users.json')
electiondb = TinyDB('election.json')
votedb = TinyDB('vote.json')

from flask_mail import Message

app = Flask(__name__)
CORS(app)

# def sendConfirmation(user):
#     msg = Message("Hello from Fengyu",
#                   sender="caifengyutruman@gmail.com",
#                   recipients=[user['email']])
#     with app.app_context():
#         mail.send(msg)

@app.route('/users/authenticate', methods=['POST'])
def authenticate():
    if request.method == 'POST':
        context = json.loads(request.data)

        # find username first
        incomingUser = context['username']
        incomingPwd = context['password']

        if not db.contains(where('username') == incomingUser):
            return Response(status=404)
        else:
            # check the correctness of password
            dbUser = db.search(where('username') == incomingUser)[0]
            if dbUser['password'] == incomingPwd:
                print('found')
                res = dict()
                res['id'] = dbUser['id']
                res['username'] = dbUser['username']
                res['firstName'] = dbUser['firstName']
                res['lastName'] = dbUser['lastName']
                res['email'] = dbUser['email']
                res['token'] = 'random'
                res = json.dumps(res)
                return Response(res, status=200)
            else:
                return Response(status=404)
    
    return Response(status=404)

@app.route('/createElection', methods=['POST'])
def createPoll():
    if request.method == 'POST':
        poll = json.loads(request.data)
        pollName = poll['name']

        if electiondb.contains(where('name') == pollName):
            return Response(status=404)
        else:
            electiondb.insert(poll)
            return Response('success', status=200)
    return Response(status=404)


@app.route('/getElection', methods=['GET'])
def getElection():
    if request.method == 'GET':
        elections = electiondb.all()
        res = json.dumps(elections)

        return Response(res, mimetype='application/json')
    
    return Response(status=404)

@app.route('/vote', methods=['POST'])
def vote():
    if request.method == 'POST':
        vote = json.loads(request.data)
        votedb.insert(vote)
        
        return Response(status=200)
    
    return Response(status=404)

@app.route('/getVoted', methods=['POST'])
def getVoted():
    if request.method == 'POST':
        myID = json.loads(request.data)
        myID = myID['voter']
        print(myID)
        alreadyVotes = votedb.search(where('voter')==myID)
        print(alreadyVotes)
        res = json.dumps(alreadyVotes)
        return Response(res, status=200)
    
    return Response(status=404)
        

@app.route('/users', methods=['GET'])
def getUsers():
    if request.method == 'GET':
        # return the data of user
        # user id first name/last name
        print('Get all users')
        allUsers = db.all()
        res = json.dumps(allUsers)
        return Response(res, status=200)
        
    return Response(status=404)

@app.route('/users/register', methods=['POST'])
def register():
    if request.method == 'POST':
        # check the duplicate

        # if ok, put it into the database
        # else return failure
        registerInfo = json.loads(request.data)
        if db.contains(where('username') == registerInfo['username']):
            return Response(status=400)
        else:
            registerInfo['id'] = len(db)
            db.insert(registerInfo)
            # sendConfirmation(registerInfo)
            return Response(status=200)

    return Response(status=404)

@app.route('/users/<id>', methods=['POST'])
def getUserInfo(id):
    if request.method == 'POST':
        # get information for one user
        print(id)
        
    return Response(status=200)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Config')
    parser.add_argument("-i", "--ip", type=str, help="Please give the ip", default="127.0.0.1")
    parser.add_argument("-p", "--port", type=int, help="Please give the running port", default=4000)
    args = parser.parse_args()
    app.logger.setLevel(logging.INFO)
    app.run(debug=True, host=args.ip, port=args.port)