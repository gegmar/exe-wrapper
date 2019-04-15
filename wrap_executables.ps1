$Domain="<ANY_AD_DOMAIN>"
$ExpiresAt="2018-05-31T23:55:55"
$PayloadDir="./payload/"
$OutputDir="./output"

$GoFileTemplate = @"
package main

// Payload holds the base 64 encoded payload-exe
var Payload string
"@

# Function for getting and base64 encoding payload content
function encodeFileB64([String] $payloadPath) {
    $fileContentEncoded = [Convert]::ToBase64String([IO.File]::ReadAllBytes($payloadPath))
    return $fileContentEncoded
}

# Prepare the app icon (icon.ico) to be embedded into the resulting .exe-files
rsrc.exe -ico .\icon.ico -o resources.syso

# Foreach raw-payload from lucy, build a new payload with our policy-wrapper
foreach( $file in Get-ChildItem $PayloadDir -Filter *.exe)
{
    $fileName = $file.Name
    $b64Payload = encodeFileB64($file.FullName)
    
    $GoFileTemplate + ' = "' + $b64Payload + '"' | Set-Content(".\payload.go")
    go build -ldflags "-H=windowsgui -X main.WindowsDomain=$Domain -X main.ExpirationDate=$ExpiresAt" -o "$OutputDir/$fileName"
}

# Restore original content of template .go-File
$GoFileTemplate | Set-Content(".\payload.go")