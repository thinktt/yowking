#!/bin/bash

cp -r /build/assets/cm/books /build/books || exit 1
cd /build/books || exit 1

declare -a books=( "FischerR.obk" "AlekhineA.obk" "BotvinnikM.obk" "Depth4.obk" "FastLose.obk" "IvanchukV.OBK" "KramnikV.OBK" "MentorFrench.obk" "OldBook.obk" "ReshevskyS.obk" "SmyslovV.obk" "TimmanJ.obk" "AnandV.obk" "CMX.obk" "Depth6.obk" "FineR.obk" "KamskyG.obk" "LarsenB.obk" "MentorGerman.OBK" "PaulsenL.OBK" "RetiR.obk" "SpasskyB.obk" "Trappee.obk" "AnderssenA.obk" "CapablancaJR.obk" "Drawish.obk" "KarpovA.obk" "LaskerE.obk" "Merlin.obk" "PawnMoves.obk" "RubinsteinA.obk" "SteinitzW.obk" "Trapper.obk" "Bipto.obk" "CaptureBook.obk" "EarlyQueen.obk" "KashdanI.OBK" "LekoP.OBK" "MorphyP.obk" "PetrosianT.obk" "SeirawanY.obk" "Strong.obk" "Unorthodox.obk" "BirdH.obk" "ChigorinM.OBK" "EuweM.obk" "FlohrS.obk" "KeresP.obk" "LowCaptureBook.obk" "NajdorfM.OBK" "PillsburyH.obk" "ShirovA.obk" "TalM.obk" "WaitzkinJ.obk" "BlackburneJ.obk" "DangerDon.obk" "EvansL.obk" "Gambit.obk" "KnightMoves.obk" "MarshallF.obk" "NimzowitschA.obk" "PolgarJ.obk" "ShortN.obk" "TarraschS.obk" "Weak.obk" "BogoljubowE.obk" "Depth2.obk" "FastBook.obk" "GellerE.OBK" "KorchnoiV.obk" "Mentor.OBK" "NoBook.obk" "Reference.OBK" "SlowBook.obk" "TartakowerS.OBK" "ZukertortJ.OBK")

for f in *.OBK; do mv "$f" "${f%.OBK}.obk"; done

book2bin() {
  book=$1
  echo "Translating $book"
  if ! wine /build/assets/tools/obk2bin.exe "$book" > "/build/logs/$book.log" 2>&1; then
    echo "ERROR: $book failed see logs /build/logs/$book.log"
    exit 1
  fi
  echo "finished $book translation"
}

# normalize books
mkdir fixed
node /build/scripts/normalizeBooks.js
mv fixed/* .
rm -rf fixed

echo "Starting translation of all obk books in parallel"
obkCount=$(ls *.obk | wc -l)
echo "$obkCount obk books to translate"

for book in "${books[@]}"
do
  book2bin "$book" &
done
wait

binCount=$(ls *.bin | wc -l)
echo "$binCount bin books created"

# move output to dist
mkdir -p /build/dist/books
mv *.bin /build/dist/books
