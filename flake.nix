{
  description = "A website that provides a random fortune";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ = {
      url = "github:a-h/templ/v0.2.707";
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
            version = "unstable-2024-06-03";
            src = self;
            vendorHash = "sha256-wIMeyKx2ho4lt9x2nl3vq3I+98O1OkXDLWOU5Ez/lCM=";

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
