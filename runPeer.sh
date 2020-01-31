pkill -f Peerster
cd Peerster
rm *.txt
go build

sleep 2

./Peerster -name A -N 3 -gossipAddr 127.0.0.1:5000 -UIPort 7080 -GuiPort 8000 -peers 127.0.0.1:5001,127.0.0.1:5002 > A.txt &
./Peerster -name B -N 3 -gossipAddr 127.0.0.1:5001 -UIPort 7081 -GuiPort 8001 -peers 127.0.0.1:5000,127.0.0.1:5002 > B.txt &
./Peerster -name C -N 3 -gossipAddr 127.0.0.1:5002 -UIPort 7082 -GuiPort 8002 -peers 127.0.0.1:5000,127.0.0.1:5001 > C.txt &

./Peerster -name D -N 3 -gossipAddr 127.0.0.1:5003 -UIPort 7083 -GuiPort 8003 -peers 127.0.0.1:5000,127.0.0.1:5001 > D.txt &

sleep 1000