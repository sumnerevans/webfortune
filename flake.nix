{
  description = "A website that provides a random fortune";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ = {
      url = "github:a-h/templ/v0.2.513";
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
            version = "unstable-2023-01-09";
            src = self;
            vendorHash = "sha256-sYdS8+XvKMxRF2OrvpP92DhObYsea0oWKzy8tKFZ8FI=";

            preBuild = ''
              ${pkgs.templ}/bin/templ generate
            '';
          };
        };
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go_1_21 gopls gotools pre-commit pkgs.templ ];
        };
      }));
}
