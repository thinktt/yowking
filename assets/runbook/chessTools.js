const Chess = require("chess.js").Chess

// Wraps chess.js with useful extras.
function create(startFen) {
  startFen = startFen || "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
  const chess = new Chess(startFen)

  const chessTools = {
    chess,
    reset,
    moveNumber,
    applyMoves,
    uciMoves,
    uci,
    getUciMoves,
    getUciMovesFromPgn,
    legalMoves,
    fen,
    move,
    undo,
    turn,
    squaresOf,
    squareOfKing,
    squareOfOpponentsKing,
    squaresOfPiece,
    coordinates,
    distance,
    manhattanDistance,
    euclideanDistance,
    otherPlayer,
    pickRandomMove,
    filterForcing,
    inCheckmate,
    inStalemate,
    materialEval,
    material,
    history: chess.history,
    ascii: chess.ascii,
    load_pgn: chess.load_pgn,
    pgn: chess.pgn,
  }

  return chessTools
  

  function reset() {
    chess.reset()
  }

  // returns the last whole move number in the game
  function moveNumber() {
    return chess.history({ verbose: true }).length / 2 
  }

  function applyMoves(moves) {
    moves.forEach(move => chess.move(move, { sloppy: true }))
  }

  function uciMoves() {
    const moves = chess.history({verbose: true})
    const uciMoves = getUciMoves(moves)
    return uciMoves
  }

  // Convert a chess.js move to a uci move
  function uci(move) {
    return move.from + move.to + (move.promotion || "")
  }

  // convert an array of chess.js moves to uci
  function getUciMoves(moves) {
    const uciMoves = []
    for (move of moves) {
      uciMoves.push(uci(move))
    }
    return uciMoves
  }

  // conversts a pgn to a uci move array
  function getUciMovesFromPgn(pgn) {
    const chess = new Chess()
    chess.load_pgn(game.pgn) 
    const uciMoves = getUciMoves(chess.history({ verbose: true }))
    return uciMoves
  }

  // Legal moves from current position.
  function  legalMoves() {
    return chess.moves({ verbose: true })
  }

  function fen() {
    return chess.fen()
  }

  function move(move) {
    chess.move(move)
  }

  function undo() {
    chess.undo()
  }

  function turn() {
    return chess.turn()
  }

  function squaresOf(colour) {
    return chess.SQUARES.filter(square => {
      const r = chess.get(square)
      return r && r.color === colour
    })
  }

  function squareOfKing() {
    return chessTools.squaresOfPiece(chess.turn(), "k")
  }

  function squareOfOpponentsKing() {
    return chessTools.squaresOfPiece(chessTools.otherPlayer(chess.turn()), "k")
  }

  function squaresOfPiece(colour, pieceType) {
    return chessTools.squaresOf(colour).find(
      square => chess.get(square).type.toLowerCase() === pieceType
    )
  }

  function coordinates(square) {
    return { 
      x: square.charCodeAt(0) - "a".charCodeAt(0) + 1, 
      y: Number(square.substring(1, 2)) 
    }
  }

  function distance(a, b) {
    return Math.max(Math.abs(a.x - b.x), Math.abs(a.y - b.y))
  }

  function manhattanDistance(a, b) {
    return Math.abs(a.x - b.x) + Math.abs(a.y - b.y)
  }

  function euclideanDistance(a, b) {
    const dx = (a.x - b.x)
    const dy = (a.y - b.y)
    return Math.sqrt(dx * dx + dy * dy)
  }

  function otherPlayer(colour) {
    return colour === "w" ? "b" : "w"
  }

  function pickRandomMove(moves) {
    return chessTools.uci(moves[Math.floor(Math.random() * moves.length)])
  }

  function filterForcing(legalMoves) {
    const mates = legalMoves.filter(move => /#/.test(move.san))
    return mates.length ? mates : legalMoves.filter(move => /\+/.test(move.san))
  }

  function inCheckmate() {
    return chess.in_checkmate()
  }

  function inStalemate() {
    return chess.in_stalemate()
  }

  function materialEval() {
    return chessTools.material("w") - chessTools.material("b")
  }

  function material(colour) {
    const valueOf = { p: 1, n: 3, b: 3, r: 6, q: 9, k: 0 }
    return chessTools.squaresOf(colour).map(
      square => valueOf[chess.get(square).type]).reduce((a, b) => a + b
    )
  }

}

module.exports = { 
  create, 
}
