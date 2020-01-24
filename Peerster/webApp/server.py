from flask import Flask
from flask_cors import CORS
from flask.json import jsonify
from flask import request
import sys
import os
import logging

logging.basicConfig(stream=sys.stderr, level=logging.DEBUG)

app = Flask(__name__)
CORS(app)
@app.route('/', methods = ['GET'])

def solve():
	
	peers = ["127.0.0.1:5001", "127.0.0.1:5002"]
	msgs = ["fwaefaesf", "fwaefaes ", "fewafwa"]

	return jsonify(Peers=peers, Rumors=msgs)

@app.route('/', methods = ['POST'])

def solvePOST():

	print("serving post")
	content = request.json
	print(content['Msg'])
	return jsonify({"result": 200})

app.run(host = "127.0.0.1", port = "8080")