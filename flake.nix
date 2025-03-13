{
  description = "A Nix-flake-based Go 1.24 development environment";

  inputs.nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1.*.tar.gz";

  outputs = { self, nixpkgs }:
    let
      goVersion = 24; # Change this to update the whole stack

      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ self.overlays.default ];
          config = {
            allowUnfree = true;
            # permittedInsecurePackages = [ "olm-3.2.16" ];
          };
        };
      });
    in
    {
      overlays.default = final: prev: {
        go = final."go_1_${toString goVersion}";
      };

      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {


          # TODO: find out how to set this conditionally for only macOS before I'm done
          LIBRARY_PATH = "/opt/homebrew/lib";
          CPATH = "/opt/homebrew/include";
          # shellHook = ''
          #   # Hack for darwin, since I could not get libolm to build in nix on darwin. 
          #   if [[ ${stdenv.hostPlatform.isDarwin} ]]; then 
          #     export LIBRARY_PATH="/opt/homebrew/lib"
          #     export CPATH="/opt/homebrew/include"
          #   fi
          # '';

          packages = with pkgs; [
            # go (version is specified by overlay)
            go

            # goimports, godoc, etc.
            gotools

            # https://github.com/golangci/golangci-lint
            golangci-lint
            golangci-lint-langserver
            gopls

            pkg-config

            # Needed for crypto
            # olm
          ];
        };
      });
    };
}
