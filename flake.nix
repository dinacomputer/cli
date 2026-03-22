{
  description = "CLI for the Dina platform";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = self.shortRev or self.dirtyShortRev or "dev";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "dina";
          inherit version;
          src = ./.;
          vendorHash = "sha256-TiOg0XL2I0KavA0s1eBVW2mmR6MZoKnnGLD6iD9iY1U=";
          subPackages = [ "cmd/dina" ];
          env.CGO_ENABLED = 0;
          ldflags = [
            "-s"
            "-w"
            "-X github.com/dinacomputer/cli/internal/cli.Version=${version}"
          ];
          meta = {
            description = "CLI for the Dina platform";
            homepage = "https://github.com/dinacomputer/cli";
            license = pkgs.lib.licenses.mit;
            mainProgram = "dina";
          };
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go go-task ];
        };
      });
}
