{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go_1_24
    pkgs.go-swag
  ];

  shellHook = ''
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    echo "✅ Go + golangci-lint environment ready!"
    echo "Go version: $(go version)"
    echo "Golangci-lint version: $(golangci-lint  --version)
    echo "Use: make build  # to build your Go project"
    echo "Use: make help  # to show all command"
    
  '';
}