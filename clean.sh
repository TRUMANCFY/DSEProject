# do the cleaning

rm client
rm indServer
rm tally

rm *.txt
rm *.json

cd Peerster
rm Peerster
rm *.txt
cd client
rm command-line-arguments

cd ../../

rm voter/command-line-arguments

# remove the frontend complilation
rm -rf web/frontend/dist/
rm -rf web/peerster/dist/
rm -rf web/indserver/dist/

rm web/backend/*.json

rm -rf web/frontend/node_modules/
rm -rf web/peerster/node_modules/
rm -rf web/indserver/node_modules/

zip -9 -r --exclude=*.git/* --exclude=*doc/* Cai_Fengyu_ProjectCode.zip .