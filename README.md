# E-Voting Application Based on Homomorphic Encryption and Decentralised Tally on Peerster
### Course Project for Decentralised System Engineering, Fall 2019, EPFL


### Author: Fengyu Cai, Liangwei Chen, Ali EI Abridi

### ! Before reading the code, we suggest to have a look on presentation slides first.

### System roles:
- There are four roles in our system:
	- Voter:
		- create election
		- participate the election with public key
		- call the end of election
		- view the result
	- Peerster (Trustee):
		- do the partial decryption
		- reach conscious
	- Tallier:
		- collect the partial decrypted vote 
		- tally the result
	- Independent server:
		- send the authentication secret to Peerster
		- generate public key
		- distribute partial private keys to the servers
- What is more, we need frontends to provide user interface in the framework of Vue, and also a light-weighted backend in the framework of Flask and database to support for the user management.

### Code Structure
```
.
├── Peerster/					# Contain the code                   
├── client.go              # voter entry
├── voter/                 # supporting code for voter
├── indServer.go            # Independent Server
├── tally.go               # Tallier    
├── server.sh              # Run the program             
└── web/                    # Frontend and Backend
│  ├── frontend/          	 # Frontend code for the voter
│  ├── backend/				 # User management backend
   ├── indServer/
│  └── peerster/
└── ...

```

- The main description of Peerster please refer to the original code for the [assignment](https://github.com/lchenbb/DecentralizedSystem)
- The main body of Blockchain is located at `Peerster/gossiper/naiveBlockchain.go`

### Installation
- User Interface
	- Voter:
		
		```
		cd web/frontend/
		npm install
		```
	
	- Independent server

		```
		cd web/indserver/
		npm install
		```
	
	- Peerster
		
		```
		cd web/peerster/
		npm install
		```

- Install external GoLang Package

	```
	go get -u go.dedis.ch/kyber
	go get -u github.com/gorilla/mux
	go get -u github.com/dedis/protobuf
	```

- Install python dependency, please refer to server.py

### Run the code
- Build and run voter. The default port is 8080
	
	```
	go build client.go
	./client -port=xxxx
	```

- Build and run tallier. The default port is 8082
	
	```
	go build tally.go
	./tally
	```
	
- Build and run independent server. The default is 8081
	
	```
	go build indServer.go
	./indServer
	```

- Run Peerster, also refer to the origin [assignment](https://github.com/lchenbb/DecentralizedSystem)
- Run backend server. It will automatically launch the database.

	```
	cd web/backend/
	python server.py
	```

- Compile GUI
	- All of the compilation result will be put to the corresponding `dist/`
	- Compile user interface
	
	```
	cd web/frontend/
	npm run build
	```
	
	- Compile independent server interface
	
	```
	cd web/indServer/
	npm run build
	``` 
	
	- Compile peetster interface
	
	```
	cd web/peerster/
	npm run build
	```

- Quite complicated, right? We have offer you the script `server.sh`, the same config as the demo video. Enjoy!


### Note
- If accidentally met with issue of Cross-Origin Resource Sharing (CORS), please  switch on the [extension](https://chrome.google.com/webstore/detail/allow-cors-access-control/lhobafahddgcelffkeicbaginigeejlf?hl=en) of CORS on your browser.
- The launch of Peerster should be earlier than independent server, as independent server will send the authentication secret to the trustees.
- Due to the network layer capacity limit, currently we cannot support the election with too many choices. However, one can check the correctness of blockchain through blockchain GUI.

### Reference
- David J. Wu. 2015. Fully homomorphic encryption: Cryptography’s holy grail. XRDS: Crossroads, The ACM Magazine for Students 21, 3 (2015), 24--29.

- ElGamal, T. A public key cryptosystem and a signature scheme based on discrete logarithms. In Advances in Cryptology, Proceedings of CRYPTO '84. G. Blakley and D. Chaum (Eds.). Springer, Berlin Heidelberg, 1985, 10--18.

- M. Blum, P. Feldman, S. Micali, "Non-interactive zero-knowledge and its applications", Proc. 20th Annu. ACM Symp. Theory Comput. (STOC’88), pp. 103-112, May 1988.

- Vue user management framework https://github.com/cornflourblue/vue-vuex-registration-login-example

- Helios https://github.com/google/pyrios


### Lisence
MIT
	
