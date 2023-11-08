{
  description = "A website that provides a random fortune";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs { system = system; };
        in
        {
          packages = rec {
            default = webfortune;
            webfortune = pkgs.buildGoModule rec {
              pname = "webfortune";
              version = "unstable-2023-11-07";
              src = self;

              propagatedBuildInputs = [ pkgs.olm ];

              vendorSha256 = "sha256-ivI7/eJo/mhemO6DHhh4jGthsZI1NnE4KdsW4+07lHk=";
            };
          };
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_21
              gopls
              gotools
              pre-commit
            ];
          };
        }
      ));
}
