// This is a one time script for building out the personalities.json files from the 
// personalites.cfg and the CMP files

const fs = require('fs');

arg = process.argv[2] || ''
const name = (arg.charAt(0).toUpperCase() + arg.slice(1))
console.log(process.cwd())

cmpPath = `${process.cwd()}/assets/cm/personalities`
cpmCfgPath = `${process.cwd()}/assets/personalities.cfg`
outputPath = `${process.cwd()}/dist/personalities.json`



const wizBio = 'Ye Old Wizard simulates and pays tribute to the classic Chessmaster personalities. Chessmaster was the most popular PC chess program ever made, playing and teaching chess with millions of kids and adults from 1988 to 2007. It used Johan de Köning\'s chess engine The King to simulate chess opponentes of all rating levels. Play the Wizard himself or his many chess personalities here.'
const wizStyle = "The Wizard is the top opponent of Ye Old Wizard. He will do his very best to grind you into the ground with his kindly instructive presence."
const wizSummary = "Beats puny humans"

console.log('Building personalites.json file')
run()
async function run() {
  const cmps = parseCmpCfg()
  for (const name in cmps) {
    const cmpFileVals = await parseCmpFile(name)
    cmps[name] = { ...cmps[name], ...cmpFileVals }
    // replace cpm opptrophe sub char with actual opostrophe
    cmps[name].bio = cmps[name].bio.replace('', "'")
    debrand(cmps[name])
  }
  // console.log(await parseCmpFile(name))
  // console.log(cmps[name])
  console.log(`${Object.keys(cmps).length} Personlaities for personalities.json file`)
  cmps['Wizard'] = cmps['Chessmaster']
  cmps['Wizard'].bio  = wizBio
  cmps['Wizard'].style = wizStyle
  cmps['Wizard'].summary = wizSummary
  cmps['Wizard'].name = 'Wizard'
  delete cmps['Chessmaster']

  // fix some incorrect book names from the cmp files
  cmps['Shakespeare'].book = 'PawnMoves.bin'
  cmps['Smyslov'].book = 'SmyslovV.bin'
  cmps['Shirov'].book = 'ShirovA.bin'
  
  fs.writeFileSync(outputPath, JSON.stringify(cmps, null, 2))
}


function debrand(cmp) {
  cmp.version = cmp.version.replace('Chessmaster ', '')
  cmp.style = cmp.style.replace(' in Chessmaster® Grandmaster Edition', '')
  cmp.bio = cmp.bio.replace(' (see this Classic Game in the Chessmaster Library)', '')
  cmp.bio = cmp.bio.replace(' Grandmaster Evans is a frequent contributor to Chessmaster.', '')
  cmp.style = cmp.style.replace(
    'all the Chessmaster® Grandmaster Edition opponents',
    'all the Wizard opponents'
    )
    cmp.style = cmp.style.replace('Chessmaster® Grandmaster Edition', 'Ye Old Wizard')
    cmp.bio = cmp.bio.replace(/Chessmaster/g, 'Wizard')
    cmp.style = cmp.style.replace(/Chessmaster/g, 'Wizard')
    cmp.bio = cmp.bio.replace(
      'and buys the best and latest version of Wizard as soon as it hits the stores.', 
      'and even made a web app that simulates the chess personalites of Chessmaster, his favorite old chess program.'
    )
}

async function parseCmpFile(name) {
  
  try {
    const cmp = await getPersonality(name)
    return cmp
  } catch (err) {
    // console.log('Unable to open personality file for ' + name)
    return {}
  }
}

// Fully parese the raw engine strings in personalities.cfg
// into individual personalities
function parseCmpCfg() {
  const cmps = {}
  let cmpStrings = fs.readFileSync(cpmCfgPath, 'utf8')
  cmpStrings = cmpStrings.split('\n\n')
  cmpStrings.forEach((cmpStr) => {
    const params = parseEngStrings(cmpStr)
    cmps[params.name] = params
  })
  return cmps
}

// A big mess of parsing all the engine strings
function parseEngStrings(engStrings) {
  const personality = {out : {}}
  engStrings = engStrings.split('\n')
  personality.name = engStrings[0]
  personality.ponder = engStrings[7]
  let params = engStrings.slice(1, 7).join(' ').split(' ')
  params = params.filter(param => param != 'cm_parm')
  params.forEach((param) => {
    const paramPair = param.split('=')
    personality.out[paramPair[0]] = paramPair[1]
  })
  return personality
}


function getPersonality(name) {
  const filePromise = new Promise((resolve, reject) => {
    fs.readFile(`${cmpPath}/${name}.CMP`, (err, data) => { 
      if (err) {
        console.log(err.message)
        reject(err)
        return 
      }
        const cmp = {}
        cmp.version = ab2str(data.buffer.slice(0,32))
        cmp.book = ab2str(data.buffer.slice(192, 453))
        cmp.book = cmp.book.replace('.OBK', '.bin').replace('.obk', '.bin')
        cmp.face = ab2str(data.buffer.slice(452, 482))
        cmp.face = cmp.face.replace('.BMP', '.png').replace('.bmp', '.png')
        cmp.summary = ab2str(data.buffer.slice(482, 582))
        cmp.bio = ab2str(data.buffer.slice(582, 1581))
        cmp.raw = new Int32Array(data.buffer.slice(32,192))
        cmp.raw = Array.from(cmp.raw)
        cmp.rating = cmp.raw[6]
        cmp.style = ab2str(data.buffer.slice(1582)).replace('%d', cmp.rating)
        resolve(cmp)
    })
  })
  return filePromise
}


function printRawParams(params) {
  for (let i=0; i < params.length; i++) {
    process.stdout.write(`${params[i]} `);
    if ((i + 1) % 4 == 0 ) process.stdout.write('\n')
    // if ((i + 1) % 16 == 0 ) process.stdout.write('\n')
  }
}

function getIntsFromBuff(buffArr) {
  let ints = []
  for (let i=0; i < buffArr.byteLength; i++) {
    // console.log(buffArr.readInt32LE(i))
    ints.push(buffArr.readInt32LE(i))
  }
  return ints
}

function ab2strRaw(buf) {
  return String.fromCharCode.apply(null, new Uint8Array(buf));
}

function ab2str(buf) {
  let charCodes =  new Uint8Array(buf)
  let str = ''
  for (code of charCodes) {
    if (code == 0) break
    let char = String.fromCharCode(code)
    str = str.concat(char)
  }
  return str
}