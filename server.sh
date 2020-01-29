# kill process
kill -9 $(lsof -i:4000 -t)
kill -9 $(lsof -i:8082 -t)
kill -9 $(lsof -i:8081 -t)
kill -9 $(lsof -i:8080 -t)
kill -9 $(lsof -i:8079 -t)
kill -9 $(lsof -i:8078 -t)

# compile the frontend
cd web/frontend/
npm run build

cd ../indserver/
npm run build

cd ../backend/
rm election.json
rm vote.json
python server.py > pyserver.txt &

cd ../../

# build tally
go build tally.go
./tally > tally.txt &

# build independent server
go build indServer.go
./indServer > indServer.txt &

sh ./runPeer.sh &

# run clients
go build client.go
./client -port=8078 &
./client -port=8079 &
./client &