{
  description = "A simple Go package";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs";

  outputs = { self, nixpkgs }:
    let

      # to work with older version of flakes
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";

      # Generate a user-friendly version number.
      version = builtins.substring 0 8 lastModifiedDate;

      # System types to support.
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });

    in
    {

      # Provide some binary packages for selected system types.
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        rec {
          gopass-double-banger = pkgs.buildGoModule {
            pname = "gopass-double-banger";
            inherit version;
            # In 'nix develop', we don't need a copy of the source tree
            # in the Nix store.
            src = ./.;

            # This hash locks the dependencies of this package. It is
            # necessary because of how Go requires network access to resolve
            # VCS.  See https://www.tweag.io/blog/2021-03-04-gomod2nix/ for
            # details. Normally one can build with a fake sha256 and rely on native Go
            # mechanisms to tell you what the hash should be or determine what
            # it should be "out-of-band" with other tooling (eg. gomod2nix).
            # To begin with it is recommended to set this, but one must
            # remember to bump this hash when your dependencies change.
            # vendorHash = pkgs.lib.fakeHash;

            vendorHash = "sha256-WBchbX6b/mkz7uP+v3ULJIlFYcwECp8AxAqtS4xLdAg=";
          };
          bootstrap-yk-oath = pkgs.writeShellApplication {
            name = "bootstrap-yk-oath";
            runtimeInputs = [
              gopass-double-banger
            ];
            text = ''
              [ $# -eq 2 ] || { echo "Usage: $0 <YK_DEVICE> <TOTP_PATH>"; exit 1; }
              YK_DEVICE=$1
              TOTP_PATH=$2
              [ -z "$(gopass ls -f "$TOTP_PATH")" ] && echo "No TOTP entries found in $TOTP_PATH" && exit 1
              ykman --device "$YK_DEVICE" oath reset
              xargs -a <(gopass ls -f "$TOTP_PATH") -d '\n' gopass-double-banger show | xargs -n1 -d '\n' ykman --device "$YK_DEVICE" oath accounts uri -f -t
              ykman --device "$YK_DEVICE" oath access change
            '';
          };
          capture-totp-qr = pkgs.writeShellApplication {
            name = "capture-totp-qr";
            runtimeInputs = [
              gopass-double-banger
              pkgs.imagemagick
              pkgs.zbar
            ];
            text = ''
              [ $# -eq 1 ] || { echo "Usage: $0 <TOTP_PATH>"; exit 1; }
              TOTP_PATH=$1
              TOTP_URI=$(import -silent -window root bmp:- | zbarimg - | sed -e 's/QR-Code://')
              echo "$TOTP_URI" | grep -q '^otpauth://totp/' || { echo "No QR code found"; exit 1; }
              gopass-double-banger insert "$TOTP_PATH" <(echo "$TOTP_URI")
            '';
          };
        });
      
      # Add dependencies that are only needed for development
      devShells = forAllSystems (system:
        let 
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [ go gopls gotools go-tools gnumake ];
          };
        });

      # The default package for 'nix build'. This makes sense if the
      # flake provides only one package or there is a clear "main"
      # package.
      defaultPackage = forAllSystems (system: self.packages.${system}.gopass-double-banger);
    };
}
