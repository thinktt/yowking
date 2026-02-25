const book = require('./book2.js')

const jsonStr = process.argv[2]

let input
try {
  input = JSON.parse(jsonStr)
} catch (e) {
  process.stderr.write(`error parsing json input: ${e}`)
  process.exit(1)
}

const fen = input.fen
const bookName = input.book
const booksDir = input.booksDir

if (!fen || !bookName) {
  process.stderr.write('expected input JSON with "fen" and "book"')
  process.exit(1)
}

dumpMoves(fen, bookName, booksDir)

async function dumpMoves(fen, bookName, booksDir) {
  let err = null
  const bookMoves = await book.getAllBookMoves(fen, bookName, { booksDir }).catch(e => err = e)
  if (err) {
    process.stderr.write(`error getting book moves: ${err.message}`)
    process.exit(1)
  }

  // Emit only stable fields needed for comparison.
  const out = (bookMoves || []).map(m => ({
    move: m._algebraic_move,
    weight: m._weight,
    learn: m._learn,
  }))

  process.stdout.write(JSON.stringify(out))
  process.exit(0)
}
