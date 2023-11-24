{
  description = "A website that provides a random fortune";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ = {
      url = "github:a-h/templ";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, templ }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs { inherit system; };
          templ-pkg = templ.packages.${system}.templ;
        in
        {
          packages = rec {
            default = webfortune;
            webfortune = pkgs.buildGoModule rec {
              pname = "webfortune";
              version = "unstable-2023-11-07";
              src = self;
              vendorHash = "sha256-njI7D0eOq4QELwG14xndPcG17ZJzxzrTpBUnjTWuTyw=";

              propagatedBuildInputs = [ pkgs.olm ];

              preBuild = ''
                ${templ-pkg}/bin/templ generate
              '';
            };
          };
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_21
              gopls
              gotools
              pre-commit
              templ-pkg
            ];
          };
        }
      ));
}
