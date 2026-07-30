#!/bin/bash
# Download popular Google Fonts (Regular weight TTF files)
# Source: Google Fonts GitHub repo

BASE="https://raw.githubusercontent.com/google/fonts/main"

declare -A FONTS=(
  # Sans-Serif
  ["Roboto"]="ofl/roboto/Roboto%5Bwdth%2Cwght%5D.ttf"
  ["Open Sans"]="ofl/opensans/OpenSans%5Bwdth%2Cwght%5D.ttf"
  ["Lato"]="ofl/lato/Lato-Regular.ttf"
  ["Montserrat"]="ofl/montserrat/Montserrat%5Bwght%5D.ttf"
  ["Poppins"]="ofl/poppins/Poppins-Regular.ttf"
  ["Nunito"]="ofl/nunito/Nunito%5Bwght%5D.ttf"
  ["Raleway"]="ofl/raleway/Raleway%5Bwght%5D.ttf"
  ["Inter"]="ofl/inter/Inter%5Bopsz%2Cwght%5D.ttf"
  ["Ubuntu"]="ufl/ubuntu/Ubuntu-Regular.ttf"
  ["Rubik"]="ofl/rubik/Rubik%5Bwght%5D.ttf"
  ["Work Sans"]="ofl/worksans/WorkSans%5Bwght%5D.ttf"
  ["Nunito Sans"]="ofl/nunitosans/NunitoSans%5Bopsz%2Cwdth%2Cwght%5D.ttf"
  ["Fira Sans"]="ofl/firasans/FiraSans-Regular.ttf"
  ["Barlow"]="ofl/barlow/Barlow-Regular.ttf"
  ["Mulish"]="ofl/mulish/Mulish%5Bwght%5D.ttf"
  ["Cabin"]="ofl/cabin/Cabin%5Bwdth%2Cwght%5D.ttf"
  ["Karla"]="ofl/karla/Karla%5Bwght%5D.ttf"
  ["Manrope"]="ofl/manrope/Manrope%5Bwght%5D.ttf"
  ["Quicksand"]="ofl/quicksand/Quicksand%5Bwght%5D.ttf"
  ["Josefin Sans"]="ofl/josefinsans/JosefinSans%5Bwght%5D.ttf"
  ["DM Sans"]="ofl/dmsans/DMSans%5Bopsz%2Cwght%5D.ttf"
  ["Noto Sans"]="ofl/notosans/NotoSans%5Bwdth%2Cwght%5D.ttf"
  ["Source Sans 3"]="ofl/sourcesans3/SourceSans3%5Bwght%5D.ttf"
  ["Outfit"]="ofl/outfit/Outfit%5Bwght%5D.ttf"
  ["Plus Jakarta Sans"]="ofl/plusjakartasans/PlusJakartaSans%5Bwght%5D.ttf"
  ["Albert Sans"]="ofl/albertsans/AlbertSans%5Bwght%5D.ttf"
  ["Lexend"]="ofl/lexend/Lexend%5Bwght%5D.ttf"
  ["Red Hat Display"]="ofl/redhatdisplay/RedHatDisplay%5Bwght%5D.ttf"
  ["Sora"]="ofl/sora/Sora%5Bwght%5D.ttf"
  ["Space Grotesk"]="ofl/spacegrotesk/SpaceGrotesk%5Bwght%5D.ttf"
  ["Figtree"]="ofl/figtree/Figtree%5Bwght%5D.ttf"
  ["Overpass"]="ofl/overpass/Overpass%5Bwght%5D.ttf"
  ["Exo 2"]="ofl/exo2/Exo2%5Bwght%5D.ttf"
  ["Asap"]="ofl/asap/Asap%5Bwdth%2Cwght%5D.ttf"
  ["Public Sans"]="ofl/publicsans/PublicSans%5Bwght%5D.ttf"
  ["Jost"]="ofl/jost/Jost%5Bwght%5D.ttf"
  ["Archivo"]="ofl/archivo/Archivo%5Bwdth%2Cwght%5D.ttf"
  ["Hind"]="ofl/hind/Hind-Regular.ttf"
  ["Signika"]="ofl/signika/Signika%5BGRAD%2Cwght%5D.ttf"
  ["Varela Round"]="ofl/varelaround/VarelaRound-Regular.ttf"
  ["Catamaran"]="ofl/catamaran/Catamaran%5Bwght%5D.ttf"
  ["Dosis"]="ofl/dosis/Dosis%5Bwght%5D.ttf"
  ["Maven Pro"]="ofl/mavenpro/MavenPro%5Bwght%5D.ttf"
  ["Sarabun"]="ofl/sarabun/Sarabun-Regular.ttf"
  ["Titillium Web"]="ofl/titilliumweb/TitilliumWeb-Regular.ttf"
  ["Assistant"]="ofl/assistant/Assistant%5Bwght%5D.ttf"
  ["Kanit"]="ofl/kanit/Kanit-Regular.ttf"

  # Serif
  ["Merriweather"]="ofl/merriweather/Merriweather-Regular.ttf"
  ["Playfair Display"]="ofl/playfairdisplay/PlayfairDisplay%5Bwght%5D.ttf"
  ["Lora"]="ofl/lora/Lora%5Bwght%5D.ttf"
  ["PT Serif"]="ofl/ptserif/PTSerif-Regular.ttf"
  ["Noto Serif"]="ofl/notoserif/NotoSerif%5Bwdth%2Cwght%5D.ttf"
  ["Libre Baskerville"]="ofl/librebaskerville/LibreBaskerville-Regular.ttf"
  ["Source Serif 4"]="ofl/sourceserif4/SourceSerif4%5Bopsz%2Cwght%5D.ttf"
  ["EB Garamond"]="ofl/ebgaramond/EBGaramond%5Bwght%5D.ttf"
  ["Crimson Text"]="ofl/crimsontext/CrimsonText-Regular.ttf"
  ["Bitter"]="ofl/bitter/Bitter%5Bwght%5D.ttf"
  ["Cormorant Garamond"]="ofl/cormorantgaramond/CormorantGaramond-Regular.ttf"
  ["Spectral"]="ofl/spectral/Spectral-Regular.ttf"
  ["DM Serif Display"]="ofl/dmserifdisplay/DMSerifDisplay-Regular.ttf"
  ["Vollkorn"]="ofl/vollkorn/Vollkorn%5Bwght%5D.ttf"
  ["Alegreya"]="ofl/alegreya/Alegreya%5Bwght%5D.ttf"
  ["Cardo"]="ofl/cardo/Cardo-Regular.ttf"
  ["Libre Caslon Text"]="ofl/librecaslontext/LibreCaslonText-Regular.ttf"
  ["Bodoni Moda"]="ofl/bodonimoda/BodoniModa%5Bopsz%2Cwght%5D.ttf"
  ["Fraunces"]="ofl/fraunces/Fraunces%5Bopsz%2Cwght%5D.ttf"

  # Monospace
  ["Roboto Mono"]="ofl/robotomono/RobotoMono%5Bwght%5D.ttf"
  ["Source Code Pro"]="ofl/sourcecodepro/SourceCodePro%5Bwght%5D.ttf"
  ["Fira Code"]="ofl/firacode/FiraCode%5Bwght%5D.ttf"
  ["JetBrains Mono"]="ofl/jetbrainsmono/JetBrainsMono%5Bwght%5D.ttf"
  ["IBM Plex Mono"]="ofl/ibmplexmono/IBMPlexMono-Regular.ttf"
  ["Space Mono"]="ofl/spacemono/SpaceMono-Regular.ttf"
  ["Inconsolata"]="ofl/inconsolata/Inconsolata%5Bwdth%2Cwght%5D.ttf"

  # Display / Headings
  ["Oswald"]="ofl/oswald/Oswald%5Bwght%5D.ttf"
  ["Bebas Neue"]="ofl/bebasneue/BebasNeue-Regular.ttf"
  ["Anton"]="ofl/anton/Anton-Regular.ttf"
  ["Abril Fatface"]="ofl/abrilfatface/AbrilFatface-Regular.ttf"
  ["Righteous"]="ofl/righteous/Righteous-Regular.ttf"
  ["Teko"]="ofl/teko/Teko%5Bwght%5D.ttf"
  ["Fjalla One"]="ofl/fjallaone/FjallaOne-Regular.ttf"
  ["Staatliches"]="ofl/staatliches/Staatliches-Regular.ttf"
  ["Russo One"]="ofl/russoone/RussoOne-Regular.ttf"
  ["Bungee"]="ofl/bungee/Bungee-Regular.ttf"

  # Handwriting / Script
  ["Dancing Script"]="ofl/dancingscript/DancingScript%5Bwght%5D.ttf"
  ["Pacifico"]="ofl/pacifico/Pacifico-Regular.ttf"
  ["Satisfy"]="ofl/satisfy/Satisfy-Regular.ttf"
  ["Great Vibes"]="ofl/greatvibes/GreatVibes-Regular.ttf"
  ["Sacramento"]="ofl/sacramento/Sacramento-Regular.ttf"
  ["Caveat"]="ofl/caveat/Caveat%5Bwght%5D.ttf"
  ["Kalam"]="ofl/kalam/Kalam-Regular.ttf"
  ["Patrick Hand"]="ofl/patrickhand/PatrickHand-Regular.ttf"
  ["Indie Flower"]="ofl/indieflower/IndieFlower-Regular.ttf"
  ["Shadows Into Light"]="ofl/shadowsintolight/ShadowsIntoLight-Regular.ttf"
)

total=${#FONTS[@]}
count=0
failed=0

for name in "${!FONTS[@]}"; do
  count=$((count + 1))
  # Convert name to filename-safe format
  filename=$(echo "$name" | tr ' ' '_')
  url="${BASE}/${FONTS[$name]}"

  printf "[%3d/%d] Downloading %s... " "$count" "$total" "$name"

  if curl -sL --fail -o "${filename}.ttf" "$url" 2>/dev/null; then
    size=$(wc -c < "${filename}.ttf" 2>/dev/null)
    if [ "$size" -gt 1000 ]; then
      echo "OK ($(( size / 1024 ))KB)"
    else
      echo "FAILED (too small)"
      rm -f "${filename}.ttf"
      failed=$((failed + 1))
    fi
  else
    echo "FAILED"
    rm -f "${filename}.ttf"
    failed=$((failed + 1))
  fi
done

echo ""
echo "Done: $((count - failed)) fonts downloaded, $failed failed"
