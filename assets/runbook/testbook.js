const chessTools = require("./chessTools.js")
const chalk = require('chalk')
const book = require('./book2')
const personalites = require('./personalities.js')


// const moves = [ 'd2d4', 'c7c5', 'c2c4', 'c5d4' ]
const moves = []

getABunchOfMoves()

async function getABunchOfMoves() {
  const bookMoves = {}
  for (let i = 0; i<1000; i++) {
    let bookMove = await getNextMove(moves, 'JW7', 'xyzzy')

    if (bookMove === undefined) bookMove = {move :'none'}
  
    if (!bookMoves[bookMove.move]) {
      bookMoves[bookMove.move] = 1
    } else {
      bookMoves[bookMove.move]++
    }
  
  }
  console.log(bookMoves)
}


async function getNextMove(moves, wizPlayer, gameId) {
      
  if (!wizPlayer) {
    console.log(`No personality selected for ${gameId}, no move made`)
    return
  }

  const cmp = personalites.getSettings(wizPlayer)

  console.log(chalk.blue(`Moving as ${cmp.name}`)) 
  console.log(chalk.blue(`Using ${cmp.book} for book moves`))


  const chess = chessTools.create()
  chess.applyMoves(moves)
  const legalMoves = chess.legalMoves()
  
  // if there are no legal moves then return
  if (!legalMoves.length) {
    return
  }

  const bookMove = await book.getHeavyMove(chess.fen(), cmp.book)
  // const bookMove = await book.getRandomMove(chess.fen())
  // const bookMove = ''

  if (bookMove != "") {
    console.log(`bookMove: ${bookMove}`)
    return {move: bookMove, willAcceptDraw: false}
  }

}