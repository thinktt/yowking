import { writeFileSync } from 'fs'

const cpuStart = parseInt(process.env.CPU_START)
const cpuEnd = parseInt(process.env.CPU_END)
if (!cpuStart || !cpuEnd) {
  console.log('env CPU_START and CPU_END must be set')
  process.exit(1)
} else {
 console.log('using cpuStart', cpuStart, 'and cpuEnd', cpuEnd)
}

const cpuSets = []

// make a cpuSet list from the start and end
for (let i = cpuStart; i <= cpuEnd; i++) {
  cpuSets.push(i)
}

console.log('using cpuSets', cpuSets)

let data = `version: '3'\nservices:`
for (let i = 0; i < cpuSets.length; i+=2) {
  const cpu1 = cpuSets[i]
  const cpu2 = cpuSets[i+1]
  data += `
    yowking${cpu1}${cpu2}:
      image: zen:5000/yowking
      cpus: 2
      cpuset: "${cpu1},${cpu2}"
      volumes:
        - cal45:/opt/yowking/calibrations
      env_file: 
        - env/king.env
      container_name: yowking${cpu1}${cpu2}
      networks:
        - yow
      restart: always`
}

data += `\n\nnetworks:\n  yow:\n    external: true\n`;
data += `\nvolumes:\n  cal45:\n    external: true\n`;

writeFileSync('compose-yowking.yaml', data, 'utf8')