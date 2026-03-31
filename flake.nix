{
  description = "CLI for the Dina platform";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      releasedVersion = "0.0.3";

      platformMap = {
        x86_64-linux = "linux_amd64";
        aarch64-linux = "linux_arm64";
        x86_64-darwin = "darwin_amd64";
        aarch64-darwin = "darwin_arm64";
      };

      releaseHashes = {
        linux_amd64 = "sha256-9A6tPMUNLmGCW/kvQLxud9ApKtdu0E6B7MfU77xOoDQ=";
        linux_arm64 = "sha256-LukXsu7TgoqCe4ma3l62cqWfseQRkqJuyAvUDEcF+eI=";
        darwin_amd64 = "sha256-odgl43E4zlN49P9ft7PLBY1ig4wjQgtvQ2uFQaYtKJk=";
        darwin_arm64 = "sha256-T75HtuIw6/9QlQ3ikNJOy3TbaY2OUKzNtZE3+/RgNB0=";
      };
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = self.shortRev or self.dirtyShortRev or "dev";
        platform = platformMap.${system} or (throw "Unsupported system: ${system}");
        tarball = "dina_${releasedVersion}_${platform}.tar.gz";
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

        packages.released = pkgs.stdenvNoCC.mkDerivation {
          pname = "dina";
          version = releasedVersion;
          src = pkgs.fetchurl {
            url = "https://github.com/dinacomputer/cli/releases/download/v${releasedVersion}/${tarball}";
            hash = releaseHashes.${platform};
          };
          sourceRoot = ".";
          unpackPhase = "tar xzf $src";
          installPhase = ''
            install -Dm755 dina $out/bin/dina
          '';
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
