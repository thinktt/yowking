# The King Chess Service
The King service turns Johan de Koning's chess engine [The King](https://www.chessprogramming.org/The_King), used by the the [Chessmaster](https://en.wikipedia.org/wiki/Chessmaster) game series, into a web service. 

Currently the main component in this repo is the Yow Worker. This worker connects to a [Nats](https://nats.io/) stream and receives and replises to chess move request given a move and a Personalitiy to play the move as. 

Currently this service is used by the [Ye Old Wizard Chess Bot](https://github.com/thinktt/yeoldwiz/tree/master/yowbot)

An API for this will be finished soon that will directly make calls to the Nats streams and send request to the  Yow Workers.

## What you'll need
This setup is still highly dependent on the main Ye Old Wizard services. 
1. clone [yeoldwiz repo](https://github.com/thinktt/yeoldwiz) in a folder beside this one
2. Build yowdep dependencies in yeolwiz repo (see repo)
3. Set up your .env (see env-template)
4. Start Yow Bot and Nats (see yeoldwiz repo)
5. Start this the Yow Worker
