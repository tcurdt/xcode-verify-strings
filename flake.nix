{
  description = "xcode-verify-strings - verify Xcode .strings files";

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
            vendorHash = "sha256-qHI3iv9Xa1rz6JOMrdN3NXjJBe/ocBfoQGmiRl+YTrc=";

            postInstall = ''
              mv $out/bin/v1 $out/bin/xcode-verify-strings
            '';

            meta = with pkgs.lib; {
              description = "xcode-verify-strings - verify Xcode .strings files";
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
