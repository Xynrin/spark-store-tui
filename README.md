# Spark Store TUI APT Repository

Install:

`ash
sudo install -d -m 0755 /etc/apt/sources.list.d
printf '%s\n' 'deb [trusted=yes] https://xynrin.github.io/spark-store-tui stable main' | sudo tee /etc/apt/sources.list.d/spark-store-tui.list
sudo apt update
sudo apt install spark-store-tui
`

Package: spark-store-tui 0.7.2-1

Note: this repository is currently unsigned, so the source line uses [trusted=yes].