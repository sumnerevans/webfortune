{
  description = "A website that provides a random fortune";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ = {
      url = "github:a-h/templ/v0.2.778";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, templ }:
    (flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays =
            [ (final: prev: { inherit (templ.packages.${system}) templ; }) ];
        };
      in {
        packages = rec {
          default = webfortune;
          webfortune = pkgs.buildGoModule {
            pname = "webfortune";
            version = "unstable-2024-10-23";
            src = self;
            vendorHash = "sha256-f+uSSsjNyAVT2hLUgP6CDQrh3GvoJX53/BTMUdPzVzM=";

            preBuild = ''
              ${pkgs.templ}/bin/templ generate
            '';
          };
        };
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go gopls gotools pre-commit pkgs.templ ];
        };
      }));
}
