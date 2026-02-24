const fs = require('fs')

// Function to read a file and parse it into an object
function parseFile(filePath) {
  const content = fs.readFileSync(filePath, 'utf-8')
  const lines = content.trim().split('\n')
  const obj = {}

  for (const line of lines) {
    const [hash, path] = line.split('  ') // Note the double space
    obj[hash] = path
  }

  return obj
}

// Parse the two files into objects
const sum1 = parseFile('sum1.txt')
const sum2 = parseFile('sum2.txt')

// Compare the hashes and find ones that don't match
const nonMatchingFile = []

for (const hash in sum1) {
  if (sum2[hash]) continue
  nonMatchingFile.push(sum1[hash])
}

for (const file of nonMatchingFile) {
  console.log(file)
}
