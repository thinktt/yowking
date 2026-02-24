// Some of the Books in CM11 are not compatible with obk2bin.exe (see tools). 
// They can be fixed by replaceing the code at the beginnigg of the binary
// this script does that. 

const fs = require('fs')

const books = ['Bipto.obk','CaptureBook.obk','DangerDon.obk','Depth2.obk','Depth4.obk','Depth6.obk','Drawish.obk','EarlyQueen.obk','FastBook.obk','FastLose.obk','Gambit.obk','KnightMoves.obk','LowCaptureBook.obk', 'Merlin.obk','NoBook.obk','OldBook.obk','PawnMoves.obk','SlowBook.obk','Strong.obk','Trappee.obk','Trapper.obk','Unorthodox.obk','Weak.obk']

console.log('Normalizing Books')

// obk folder fixed hash 1acc78250646228ef3243066ba9d1882
// final bin folder no mentor books hash 

const headerData = Buffer.from([0x42, 0x4f, 0x4f, 0x21])
const padding = Buffer.from([0x0, 0x0, 0x0, 0x0])

// console.log(headerData)

let i = 0
for (const book of books) {
  const bookData = fs.readFileSync(book)
  console.log(bookData.slice(6,20), book)

  if (bookData.slice(0,4) + '' !== 'BOO!') {
    console.log(`normalizing ${book}`)
    const newBookData =  Buffer.concat([ 
      headerData, bookData.slice(4, 6), padding, bookData.slice(6,) 
    ])
    fs.appendFileSync(`fixed/${book}`, Buffer.from(newBookData))
    i++
  }
}
console.log('books normalized', i)
