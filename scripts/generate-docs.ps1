$ErrorActionPreference = "Stop"

Remove-Item -Recurse -Force -ErrorAction SilentlyContinue completions
New-Item -ItemType Directory -Path completions | Out-Null

task build

foreach ($sh in "bash", "zsh", "fish") {
	& ./cloudflare-utils completion $sh | Out-File -Encoding utf8 -FilePath "completions/cloudflare-utils.$sh"
}
