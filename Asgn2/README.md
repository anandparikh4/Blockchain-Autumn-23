Steps to follow:
1. Add the Asgn2 directory in fabric-samples folder (adjacent to test-network)
2. Execute run.sh shell script to setup sample network, create channel and dump install chaincode on the peers
    ./run.sh launch
3. Go to ./app subdirectory:
    cd ./app
4. Start the app as "org1" or "org2":
    node app.js org1
    OR
    node app.js org2
5. Finally, to shutdown the network, use run.sh shell script in /Asgn2 folder:
    ./run.sh kill

Chaincode for both parts is in ./chaincode/chaincode.go
Application for both parts is in ./app/app.js

Authors:
20CS10007 Anand Manojkumar Parikh
20CS10060 Soni Aditya Bharatbhai
20CS10065 Subham Ghosh
20CS30016 Divyansh Vijayvergia
