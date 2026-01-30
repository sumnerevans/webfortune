{
  description = "A website that provides a random fortune";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    templ = {
      url = "github:a-h/templ/v0.3.977";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      templ,
    }:
    (flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" ];
      perSystem =
        {
          lib,
          pkgs,
          system,
          ...
        }:
        {
          _module.args.pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [ (final: prev: { inherit (templ.packages.${system}) templ; }) ];
          };

          packages = rec {
            default = webfortune;
            webfortune = pkgs.buildGoModule {
              pname = "webfortune";
              version = "unstable-2024-10-23";
              src = self;
              vendorHash = "sha256-fPkJYxLxw4KE+Bj4tTs2RBSb7Q+b7/nag0tYaqWkS1I=";

              preBuild = ''
                ${pkgs.templ}/bin/templ generate
              '';
            };
          };
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              gotools
              pre-commit
              pkgs.templ
            ];
          };
        };
    });
}
