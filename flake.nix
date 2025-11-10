{
  description = "Miniflux development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain
            go_1_24

            # Database
            postgresql

            # Build tools
            gnumake
            git

            # Optional: useful dev tools
            gotools
            gopls
            go-tools
          ];

          shellHook = ''
            echo "🚀 Miniflux development environment loaded"
            echo "Go version: $(go version)"
            echo ""
            echo "Available commands:"
            echo "  go build      - Build the project"
            echo "  go test       - Run tests"
            echo "  go run .      - Run the application"
          '';
        };
      }
    );
}
