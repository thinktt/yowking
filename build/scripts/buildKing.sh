 #!bin/bash 

# This is a regex sed hack to replace the two bytes of TheKing350.exe we need 
# to replace to switch off opk. For reference those bytes are:
# Hex line CE80 replace 0F 85 with 90 E9

regExFind="/\x24\x10\x0f\x85\x9f\x00\x00\x00"
regExReplace="/\x24\x10\x90\xe9\x9f\x00\x00\x00/"  
echo "Building Opk free King"
cp assets/cm/TheKing350.exe dist/TheKing350.exe
sed -i -e "s${regExFind}${regExReplace}" dist/TheKing350.exe
mv dist/TheKing350.exe dist/TheKing350noOpk.exe

