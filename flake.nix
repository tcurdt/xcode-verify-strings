{
  description = "xcode-verify-strings - translate Xcode .strings files";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "xcode-verify-strings";
            version = "0.0.2";
            src = ./.;

            subPackages = [ "." ];
            vendorHash = "sha256-vv/sR6x0wcPfcT2N1Y12C4r+6ge6Re6TLPq5/e7HacI=";

            postInstall = ''
              mv $out/bin/v1 $out/bin/xcode-verify-strings
            '';

            meta = with pkgs.lib; {
              description = "xcode-verify-strings - translate Xcode .strings files";
              license = licenses.asl20;
              maintainers = [ ];
            };
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
          ];
        };
      }
    );
}
