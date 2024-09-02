process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0'
const token = process.env.YOW_TOKEN 
console.log(token)

const whitePlayers = ["Capablanca", "Wizard", "Fischer"] 
// const blackPlayers = ["Tex", "Ben", "Stanley"]
const blackPlayers = ["Lacey"] 

function getRandomPlayer(players) {
  return players[Math.floor(Math.random() * players.length)]
}

async function startGame(whitePlayerId, blackPlayerId) {
  try {
    const response = await fetch('https://localhost:8443/games2', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}` // Adding the Authorization header
      },
      body: JSON.stringify({
        whitePlayer: { id: whitePlayerId, type: "cmp" },
        blackPlayer: { id: blackPlayerId, type: "cmp" }
      }),
    })

    const gameData = await response.json()
    // Call your async function here
    console.log(gameData)
    return gameData

  } catch (error) {
    console.error('Error starting game:', error)
    return
  }

 
}
async function waitForGameEnd(gameID) {
  const controller = new AbortController()
  const signal = controller.signal

  try {
    const response = await fetch(`https://localhost:8443/streams/${gameID}`, { signal })
    const reader = response.body.getReader()
    const decoder = new TextDecoder()

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      const eventStr = decoder.decode(value)
      const event = parseSSE(eventStr)
      const data = JSON.parse(event.data) 

      console.log(data)

      if (data.status) {
        console.log("closing stream ", data.id)
        controller.abort() 
        return null
      }
    }
  } catch (error) {
    return error
  }
}
async function startGames() {
  while (true) {
    const whitePlayer = getRandomPlayer(whitePlayers)
    const blackPlayer = getRandomPlayer(blackPlayers)

    console.log(`starting ${whitePlayer} vs ${blackPlayer}`)

    const game = await startGame(whitePlayer, blackPlayer)
    if (!game) break

    const result = await waitForGameEnd(game.id)
    if (result) {
      console.error(result)
      break
    }
  }
}

function parseSSE(eventString) {
  const result = {}
  eventString.trim().split('\n').forEach(line => {
    const [key, value] = line.split(/:(.+)/)
    result[key.trim()] = value.trim()
  })
  return result
}

startGames()