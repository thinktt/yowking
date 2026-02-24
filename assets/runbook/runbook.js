const chessTools = require("./chessTools.js")
const book = require('./book2.js')

const jsonStr = process.argv[2]
// console.log(`Received input: ${jsonStr}`)

let input
try {
  input = JSON.parse(jsonStr)
} catch (e) {
  process.stderr.write(`error parsing json input: ${e}`)
  process.exit(1)
}

const moves = input.moves
const bookName = input.book

getNextMove(moves, bookName)

async function getNextMove(moves, bookName) {
  // console.log(`Getting move from ${bookName}`)

  const chess = chessTools.create()
  chess.applyMoves(moves)
  
  // need a quick hack to check if there was a iliegal move
  // if there was one history will be shorter than what we put in
  if (chess.history().length != moves.length) {
    process.stderr.write("illegal move detected")
    process.exit(1)
  }

  const legalMoves = chess.legalMoves()
  
  // if there are no legal moves then return
  if (!legalMoves.length) {
    process.stderr.write("no legal moves")
    process.exit(1)
  }

  let err = null 
  const bookMove = await book.getHeavyMove(chess.fen(), bookName).catch(e => err = e)
  if (err) {
    process.stderr.write(`error getting book move: ${err.message}`)
    process.exit(1)
  }

  if (bookMove == "") {
    process.stderr.write("no book move")
    process.exit(1) 
  }

  process.stdout.write(`${bookMove}`)
  process.exit(0)
}

