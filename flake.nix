{
  description = "A website that provides a random fortune";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ = {
      url = "github:a-h/templ/v0.2.501";
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
            webfortune = pkgs.buildGoModule {
              pname = "webfortune";
              version = "unstable-2023-12-30";
              src = self;
              vendorHash = "sha256-o7E1UvE8pSDPc0Sq/aN50pVOS038TZRIlE3p9fjHxmo=";

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
