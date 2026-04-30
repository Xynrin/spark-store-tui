# Spark Store TUI APT Repository

Install:

`ash
sudo apt update
sudo apt install -y ca-certificates curl
sudo install -d -m 0755 /etc/apt/keyrings
curl -fsSL https://xynrin.github.io/spark-store-tui/spark-store-tui-archive-keyring.gpg | sudo tee /etc/apt/keyrings/spark-store-tui-archive-keyring.gpg >/dev/null
printf '%s\n' 'deb [signed-by=/etc/apt/keyrings/spark-store-tui-archive-keyring.gpg] https://xynrin.github.io/spark-store-tui stable main' | sudo tee /etc/apt/sources.list.d/spark-store-tui.list
sudo apt update
sudo apt install spark-store-tui
`

Signing key fingerprint:

`	ext
1AE6D4E7C4DB8C016F72F8C6A4D276F9CF8E57A9
`

Package: spark-store-tui 0.7.2-1