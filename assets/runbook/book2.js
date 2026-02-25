const ChessTools = require('chess-tools')
const OpeningBook = ChessTools.OpeningBooks.Polyglot
const fs = require("fs")
const path = require("path")

module.exports = { getHeavyMove, getRandomMove, getAllBookMoves }

function getRunbookDir() {
  if (process.pkg) {
    return path.dirname(process.execPath)
  }
  return __dirname
}

// get's any random move from the book
async function getRandomMove(fen, book) {
  const bookMoves = await getAllBookMoves(fen, book)
  if (!bookMoves) return ""
  move = bookMoves[Math.floor(Math.random() * bookMoves.length)]._algebraic_move
  return move
}

// get's random book move favoring higher weights and ignoring zero weights
async function getHeavyMove(fen, book) {
  let err = null
  let bookMoves = await getAllBookMoves(fen, book).catch(e => err = e)
  if (err) {
    process.stderr.write(`error getting book moves: ${err.message}`)
    process.exit(1)
  }
  if (!bookMoves) return ""

  // sort moves by weight for easier viewing when logging the final array
  bookMoves = bookMoves.sort((a, b) => (a._weight < b._weight) ? 1 : -1)

  // create a list of moves repeating moves by their weight
  const weightedMoves = []
  for (const move of bookMoves) {
    for (let i = 0; i < move._weight; i++) {
      weightedMoves.push(move._algebraic_move) 
    } 
  }
  
  // pick a random move from the weight distributed list
  move = weightedMoves[Math.floor(Math.random() * weightedMoves.length)]
  if(!move) return ""
  return move
}

async function getAllBookMoves(fen, bookName, opts = {}) {
  // console.log(bookName)
  const book = new OpeningBook()
  const booksDir = opts.booksDir || path.join(getRunbookDir(), "books")
  const bookPath = path.join(booksDir, bookName)
  const movePromise = new Promise((resolve, reject) => {
    book.on("loaded", () => {
      // console.log("book loaded")
      let moves = []
      try {
        moves = book.find(fen)
      } catch (err) {
        process.stderr.write(`Failed to get book moves: ${err.message}`)
        // reject(err)
      }
      resolve(moves)
    })
  })
  
  const bookStream = fs.createReadStream(bookPath)
  bookStream.on("error", (err) => {
    process.stderr.write(`Failed to get book: ${err.message}`)
    process.exit(1)
  })
  book.load_book(bookStream)
  return movePromise
}
